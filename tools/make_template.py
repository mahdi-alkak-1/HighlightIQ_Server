import argparse
from pathlib import Path

import cv2


def _parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Extract a kill banner template from a video frame."
    )
    parser.add_argument("--video", required=True, help="Path to the video file.")
    parser.add_argument("--time-ms", type=int, help="Timestamp in ms.")
    parser.add_argument("--out", required=True, help="Output PNG path.")
    parser.add_argument("--roi-x", type=int, default=592, help="ROI x (default 592).")
    parser.add_argument("--roi-y", type=int, default=535, help="ROI y (default 535).")
    parser.add_argument("--roi-w", type=int, default=98, help="ROI width (default 98).")
    parser.add_argument("--roi-h", type=int, default=19, help="ROI height (default 19).")
    parser.add_argument(
        "--interactive",
        action="store_true",
        help="Step through frames with keyboard to pick the exact time.",
    )
    parser.add_argument(
        "--pick-roi",
        action="store_true",
        help="Click top-left and bottom-right corners to set ROI.",
    )
    return parser.parse_args()


def _save_roi(frame, out_path: Path, x: int, y: int, w: int, h: int) -> bool:
    roi = frame[y : y + h, x : x + w]
    if roi.size == 0:
        print("ROI crop is empty. Check ROI values and video resolution.")
        return False
    out_path.parent.mkdir(parents=True, exist_ok=True)
    if not cv2.imwrite(str(out_path), roi):
        print(f"Failed to write output: {out_path}")
        return False
    print(f"Wrote template: {out_path}")
    return True


def _interactive_pick(
    cap,
    out_path: Path,
    x: int,
    y: int,
    w: int,
    h: int,
    pick_roi: bool,
) -> int:
    fps = cap.get(cv2.CAP_PROP_FPS) or 30.0
    total = int(cap.get(cv2.CAP_PROP_FRAME_COUNT) or 0)
    frame_idx = 0
    clicks = []
    last_pos = [None, None]
    current_roi = [x, y, w, h]

    print("Interactive mode:")
    print("  n = next, p = prev, j = +10, k = -10, f = +100, b = -100, s = save, q = quit")
    if pick_roi:
        print("  click 2 points to set ROI; u = set top-left, i = set bottom-right")
        print("  c = clear clicks/points")
    print("  The window shows ROI in green; time_ms is printed in console.")

    def on_mouse(event, mx, my, _flags, _param):
        if not pick_roi:
            return
        if event == cv2.EVENT_MOUSEMOVE:
            last_pos[0] = mx
            last_pos[1] = my
            return
        if event != cv2.EVENT_LBUTTONDOWN:
            return
        clicks.append((mx, my))
        if len(clicks) == 1:
            print(f"\nclicked top-left at x={mx} y={my}")
        if len(clicks) == 2:
            (x1, y1), (x2, y2) = clicks
            rx = min(x1, x2)
            ry = min(y1, y2)
            rw = max(1, abs(x2 - x1))
            rh = max(1, abs(y2 - y1))
            current_roi[0] = rx
            current_roi[1] = ry
            current_roi[2] = rw
            current_roi[3] = rh
            clicks.clear()
            print(f"\nroi_x={rx} roi_y={ry} roi_w={rw} roi_h={rh}")

    window_name = "Template Picker"
    cv2.namedWindow(window_name)
    cv2.setMouseCallback(window_name, on_mouse)

    while True:
        if frame_idx < 0:
            frame_idx = 0
        if total and frame_idx >= total:
            frame_idx = total - 1

        cap.set(cv2.CAP_PROP_POS_FRAMES, frame_idx)
        ok, frame = cap.read()
        if not ok or frame is None:
            print("Failed to read frame.")
            return 1

        time_ms = int(round((frame_idx / fps) * 1000))
        rx, ry, rw, rh = current_roi
        preview = frame.copy()
        cv2.rectangle(preview, (rx, ry), (rx + rw, ry + rh), (0, 255, 0), 1)
        if pick_roi and clicks:
            for cx, cy in clicks:
                cv2.circle(preview, (cx, cy), 4, (0, 255, 255), -1)
        if pick_roi:
            cv2.putText(
                preview,
                f"ROI x={rx} y={ry} w={rw} h={rh}",
                (10, 30),
                cv2.FONT_HERSHEY_SIMPLEX,
                0.7,
                (0, 255, 0),
                2,
            )
        cv2.imshow(window_name, preview)
        print(f"frame={frame_idx} time_ms={time_ms}", end="\r")

        key = cv2.waitKey(0) & 0xFF
        if key == ord("n"):
            frame_idx += 1
        elif key == ord("p"):
            frame_idx -= 1
        elif key == ord("j"):
            frame_idx += 10
        elif key == ord("k"):
            frame_idx -= 10
        elif key == ord("f"):
            frame_idx += 100
        elif key == ord("b"):
            frame_idx -= 100
        elif key == ord("c") and pick_roi:
            clicks.clear()
        elif key == ord("u") and pick_roi:
            if last_pos[0] is not None:
                clicks = [(last_pos[0], last_pos[1])]
                print(f"\nset top-left at x={last_pos[0]} y={last_pos[1]}")
        elif key == ord("i") and pick_roi:
            if last_pos[0] is not None and clicks:
                clicks.append((last_pos[0], last_pos[1]))
                (x1, y1), (x2, y2) = clicks
                rx = min(x1, x2)
                ry = min(y1, y2)
                rw = max(1, abs(x2 - x1))
                rh = max(1, abs(y2 - y1))
                current_roi[0] = rx
                current_roi[1] = ry
                current_roi[2] = rw
                current_roi[3] = rh
                clicks.clear()
                print(f"\nroi_x={rx} roi_y={ry} roi_w={rw} roi_h={rh}")
        elif key == ord("s"):
            print("")
            rx, ry, rw, rh = current_roi
            if _save_roi(frame, out_path, rx, ry, rw, rh):
                print(f"Saved at time_ms={time_ms}")
            else:
                return 1
        elif key == ord("q"):
            break

    cv2.destroyAllWindows()
    print("")
    return 0


def main() -> int:
    args = _parse_args()

    video_path = Path(args.video)
    out_path = Path(args.out)
    if not video_path.is_file():
        print(f"Video not found: {video_path}")
        return 1

    cap = cv2.VideoCapture(str(video_path))
    if not cap.isOpened():
        print(f"Failed to open video: {video_path}")
        return 1

    x, y, w, h = args.roi_x, args.roi_y, args.roi_w, args.roi_h

    if args.interactive:
        try:
            return _interactive_pick(cap, out_path, x, y, w, h, args.pick_roi)
        finally:
            cap.release()

    if args.time_ms is None:
        cap.release()
        print("Please provide --time-ms or use --interactive.")
        return 1

    cap.set(cv2.CAP_PROP_POS_MSEC, args.time_ms)
    ok, frame = cap.read()
    cap.release()

    if not ok or frame is None:
        print("Failed to read frame at the given timestamp.")
        return 1

    return 0 if _save_roi(frame, out_path, x, y, w, h) else 1


if __name__ == "__main__":
    raise SystemExit(main())
