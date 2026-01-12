import cv2
import numpy as np
from pathlib import Path
import os

TEMPL_DIR = Path(__file__).resolve().parent / "templates"
DEFAULT_ROI = (592, 535, 98, 19)
TEMPLATE_FILES = [(p.stem, p) for p in sorted(TEMPL_DIR.glob("*.png"))]

def _load_gray(path: Path) -> np.ndarray:
    img = cv2.imread(str(path), cv2.IMREAD_GRAYSCALE)
    if img is None:
        raise RuntimeError(f"Template not found or unreadable: {path}")
    return img

TEMPLATES = []
for name, p in TEMPLATE_FILES:
    if p.exists():
        TEMPLATES.append((name, _load_gray(p)))

if not TEMPLATES:
    raise RuntimeError(f"No template PNGs found in: {TEMPL_DIR}")

# ---------- ROI helpers ----------
def _auto_banner_roi(frame):
    """Bottom-center banner area (resolution independent)."""
    H, W = frame.shape[:2]
    x = int(W * 0.20)
    y = int(H * 0.58)
    w = int(W * 0.60)
    h = int(H * 0.32)
    return x, y, w, h

def _clamp_roi(frame, x, y, w, h):
    H, W = frame.shape[:2]
    x1 = max(0, min(W - 1, int(x)))
    y1 = max(0, min(H - 1, int(y)))
    x2 = max(0, min(W, x1 + int(w)))
    y2 = max(0, min(H, y1 + int(h)))
    return frame[y1:y2, x1:x2], (x1, y1, x2 - x1, y2 - y1)

def _best_multiscale_match(img_gray: np.ndarray, templ_gray: np.ndarray,
                           smin: float, smax: float, sstep: float):
    best_score = -1.0
    best_loc = (0, 0)
    best_size = (templ_gray.shape[1], templ_gray.shape[0])

    s = smin
    while s <= smax + 1e-9:
        th = int(round(templ_gray.shape[0] * s))
        tw = int(round(templ_gray.shape[1] * s))
        if th >= 6 and tw >= 6 and th <= img_gray.shape[0] and tw <= img_gray.shape[1]:
            t = cv2.resize(
                templ_gray, (tw, th),
                interpolation=cv2.INTER_AREA if s < 1 else cv2.INTER_CUBIC
            )
            res = cv2.matchTemplate(img_gray, t, cv2.TM_CCOEFF_NORMED)
            _, maxv, _, maxl = cv2.minMaxLoc(res)
            if float(maxv) > best_score:
                best_score = float(maxv)
                best_loc = maxl
                best_size = (tw, th)
        s += sstep

    return best_score, best_loc, best_size

def detect_banner_events(video_path: str, req, cap, return_diagnostics: bool = False) -> list[float]:
    """
    Detect banner events with optional diagnostics.
    
    If return_diagnostics=True, returns (events, diagnostics_dict) instead of just events.
    diagnostics_dict contains:
        - max_scores: list of max scores per frame
        - hit_scores: list of scores that passed threshold
        - frame_times: corresponding frame times
        - total_frames_processed: number of frames analyzed
    """
    fps = cap.get(cv2.CAP_PROP_FPS) or 30.0

    sample_fps = float(getattr(req, "sample_fps", 60.0) or 60.0)
    step = max(1, int(round(fps / sample_fps)))

    # detection tuning
    match_thr = float(getattr(req, "elim_match_threshold", 0.4))
    loc_y_max = float(getattr(req, "loc_y_max", 1.0))
    cooldown = float(getattr(req, "cooldown_seconds", 0.6))
    min_spacing = float(getattr(req, "min_spacing_seconds", 0.6))
    min_consecutive = int(getattr(req, "min_consecutive_hits", 3) or 3)

    # multiscale search range
    smin = float(getattr(req, "template_scale_min", 0.25))
    smax = float(getattr(req, "template_scale_max", 1.25))
    sstep = float(getattr(req, "template_scale_step", 0.05))

    # ROI mode
    roi_mode = str(getattr(req, "roi_mode", "manual") or "manual").lower()
    debug = bool(getattr(req, "debug", False))

    events: list[float] = []
    last_event_t = -1e9
    hit_streak = 0
    streak_armed = True

    # Diagnostics
    diagnostics = {
        "max_scores": [],
        "hit_scores": [],
        "frame_times": [],
        "total_frames_processed": 0,
        "frames_above_threshold": 0,
    }

    # mild contrast normalize (helps compression)
    clahe = cv2.createCLAHE(clipLimit=2.0, tileGridSize=(8, 8))

    f = 0
    while True:
        ok = cap.grab()
        if not ok:
            break

        if f % step == 0:
            ok2, frame = cap.retrieve()
            if not ok2 or frame is None:
                break

            # decide ROI
            if roi_mode in ("manual", "absolute"):
                rx = int(getattr(req, "roi_x", DEFAULT_ROI[0]) or DEFAULT_ROI[0])
                ry = int(getattr(req, "roi_y", DEFAULT_ROI[1]) or DEFAULT_ROI[1])
                rw = int(getattr(req, "roi_w", DEFAULT_ROI[2]) or DEFAULT_ROI[2])
                rh = int(getattr(req, "roi_h", DEFAULT_ROI[3]) or DEFAULT_ROI[3])
                if rw > 0 and rh > 0:
                    ax, ay, aw, ah = rx, ry, rw, rh
                else:
                    ax, ay, aw, ah = _auto_banner_roi(frame)
            else:
                ax, ay, aw, ah = _auto_banner_roi(frame)

            roi_img, (_, _, rw, rh) = _clamp_roi(frame, ax, ay, aw, ah)
            if roi_img.size == 0:
                f += 1
                continue

            gray = cv2.cvtColor(roi_img, cv2.COLOR_BGR2GRAY)
            gray = clahe.apply(gray)
            gray = cv2.GaussianBlur(gray, (3, 3), 0)

            t_s = f / fps
            diagnostics["total_frames_processed"] += 1

            # ---- OR / fallback logic across templates ----
            # We accept if ANY template passes threshold+location.
            best_passing = None  # (score, name, loc, size, rel_y)
            best_score_overall = -1.0

            if match_thr < 0:
                # force hit mode
                best_passing = (1.0, "forced", (0, 0), (1, 1), 0.0)
            else:
                for name, templ in TEMPLATES:
                    tgray = clahe.apply(templ)
                    tgray = cv2.GaussianBlur(tgray, (3, 3), 0)

                    score, loc, size = _best_multiscale_match(gray, tgray, smin, smax, sstep)
                    best_score_overall = max(best_score_overall, score)

                    denom = max(1, (rh - size[1]))
                    rel_y = loc[1] / denom

                    if (score >= match_thr) and (rel_y <= loc_y_max):
                        if best_passing is None or score > best_passing[0]:
                            best_passing = (score, name, loc, size, rel_y)

            # Track diagnostics
            diagnostics["max_scores"].append(best_score_overall)
            diagnostics["frame_times"].append(t_s)
            if best_score_overall >= match_thr:
                diagnostics["frames_above_threshold"] += 1
                diagnostics["hit_scores"].append((t_s, best_score_overall))

            is_hit = best_passing is not None

            if not is_hit:
                hit_streak = 0
                streak_armed = True
            else:
                hit_streak += 1
                if (
                    streak_armed
                    and hit_streak >= min_consecutive
                    and (t_s - last_event_t) >= cooldown
                ):
                    events.append(t_s)
                    last_event_t = t_s
                    streak_armed = False

                    if debug and best_passing is not None:
                        score, name, _, _, _ = best_passing
                        os.makedirs("debug_hits", exist_ok=True)
                        outp = f"debug_hits/hit_{t_s:.2f}_{name}_{score:.2f}.png"
                        cv2.imwrite(outp, roi_img)

        f += 1

    # spacing filter
    events.sort()
    filtered = []
    for t in events:
        if not filtered or (t - filtered[-1]) >= min_spacing:
            filtered.append(t)

    # Add diagnostic summary
    if diagnostics["max_scores"]:
        diagnostics["max_score"] = max(diagnostics["max_scores"])
        diagnostics["min_score"] = min(diagnostics["max_scores"])
        diagnostics["avg_score"] = sum(diagnostics["max_scores"]) / len(diagnostics["max_scores"])
        diagnostics["threshold_used"] = match_thr
    else:
        diagnostics["max_score"] = -1
        diagnostics["min_score"] = -1
        diagnostics["avg_score"] = -1
        diagnostics["threshold_used"] = match_thr

    if return_diagnostics:
        return filtered, diagnostics
    return filtered

