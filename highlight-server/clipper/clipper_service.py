from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import os
import cv2

from clipper.killbanner_detector import detect_banner_events
from clipper.clip_merger import build_merged_clips
from clipper.video_utils import get_video_end_seconds

app = FastAPI()


class DetectKillsRequest(BaseModel):
    path: str

    max_clip_seconds: int = 60
    pre_roll_seconds: int = 5
    post_roll_seconds: int = 3
    min_clip_seconds: int = 8

    sample_fps: float = 60.0
    min_spacing_seconds: float = 1.2
    max_candidates: int = 20

    merge_gap_seconds: float = 0.0
    
    # Banner detector tuning (fixed ROI for 720p)
    roi_mode: str = "manual"
    elim_match_threshold: float = 0.4      # higher = fewer false positives
    min_consecutive_hits: int = 3          # require consecutive frames to confirm

    # ROI for 1280x720 Fortnite banner crop
    roi_x: int = 592
    roi_y: int = 535
    roi_w: int = 98
    roi_h: int = 19

    cooldown_seconds: float = 1.2


class Candidate(BaseModel):
    start_ms: int
    end_ms: int
    score: float


class DetectKillsResponse(BaseModel):
    candidates: list[Candidate]
    video_end_seconds: float
    kills_detected: int


@app.post("/detect-kills", response_model=DetectKillsResponse)
def detect_kills(req: DetectKillsRequest):
    if not os.path.isfile(req.path):
        raise HTTPException(status_code=404, detail="file not found")

    cap = cv2.VideoCapture(req.path)
    if not cap.isOpened():
        raise HTTPException(status_code=400, detail="cannot open video")

    video_end_s = get_video_end_seconds(cap)

    banner_events = detect_banner_events(req.path, req, cap, return_diagnostics=False)
    cap.release()

    # Banner-only detection
    raw_events = list(banner_events)
    
    # Apply spacing filter to deduplicate nearby events
    if raw_events:
        raw_events = sorted(raw_events)
        spacing = float(req.min_spacing_seconds)
        spaced = []
        for t in raw_events:
            if not spaced or (t - spaced[-1]) >= spacing:
                spaced.append(t)
        raw_events = spaced

    candidates_dicts = build_merged_clips(raw_events, req, video_end_s)
    candidates = [Candidate(**c) for c in candidates_dicts]

    return DetectKillsResponse(
        candidates=candidates,
        video_end_seconds=video_end_s,
        kills_detected=len(raw_events),
    )
