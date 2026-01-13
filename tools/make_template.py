import argparse
from pathlib import Path

import cv2


def _parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Extract a kill banner template from a video frame."
    )
    parser.add_argument("D:\fortnite recording video to upload\MrSavage2min", required=True, help="Path to the video file.")
    parser.add_argument("--time-ms", type=int, required=True, help="Timestamp in ms.")
    parser.add_argument("c:\HighlightIQ\highlightiq-server\clipper\templates\elim_crop.png", required=True, help="Output PNG path.")
    parser.add_argument("--roi-x", type=int, default=630, help="ROI x (default 630).")
    parser.add_argument("--roi-y", type=int, default=573, help="ROI y (default 573).")
    parser.add_argument("--roi-w", type=int, default=108, help="ROI width (default 108).")
    parser.add_argument("--roi-h", type=int, default=14, help="ROI height (default 14).")
    return parser.parse_args()


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

    cap.set(cv2.CAP_PROP_POS_MSEC, args.time_ms)
    ok, frame = cap.read()
    cap.release()

    if not ok or frame is None:
        print("Failed to read frame at the given timestamp.")
        return 1

    x, y, w, h = args.roi_x, args.roi_y, args.roi_w, args.roi_h
    roi = frame[y : y + h, x : x + w]
    if roi.size == 0:
        print("ROI crop is empty. Check ROI values and video resolution.")
        return 1

    out_path.parent.mkdir(parents=True, exist_ok=True)
    if not cv2.imwrite(str(out_path), roi):
        print(f"Failed to write output: {out_path}")
        return 1

    print(f"Wrote template: {out_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
