from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from scenedetect import detect, ContentDetector
import os

app = FastAPI()

class DetectRequest(BaseModel):
    path: str
    clip_length_seconds: int = 30
    threshold: float = 27.0
    min_clip_seconds: int = 10  # avoid tiny segments

class Candidate(BaseModel):
    start_ms: int
    end_ms: int
    score: float

class DetectResponse(BaseModel):
    candidates: list[Candidate]
    video_end_seconds: float
    scenes_detected: int

@app.post("/detect-candidates", response_model=DetectResponse)
def detect_candidates(req: DetectRequest):
    if not os.path.isfile(req.path):
        raise HTTPException(status_code=404, detail="file not found")

    if req.clip_length_seconds <= 0:
        raise HTTPException(status_code=400, detail="clip_length_seconds must be > 0")

    # PySceneDetect high-level API: returns list of (start_timecode, end_timecode) pairs. :contentReference[oaicite:2]{index=2}
    scene_list = detect(req.path, ContentDetector(threshold=req.threshold))

    if not scene_list:
        # If no scenes detected, fallback to a single candidate from 0..clip_length (or to end)
        # We’ll treat video_end as clip_length in this fallback.
        end_s = float(req.clip_length_seconds)
        return DetectResponse(
            candidates=[Candidate(start_ms=0, end_ms=int(end_s * 1000), score=0.1)],
            video_end_seconds=end_s,
            scenes_detected=0,
        )

    # Video end = end time of last scene
    last_end_tc = scene_list[-1][1]
    video_end_s = float(last_end_tc.get_seconds())

    candidates: list[Candidate] = []
    used_starts = set()

    for (start_tc, _end_tc) in scene_list:
        start_s = float(start_tc.get_seconds())
        # candidate = 30s window starting at scene start
        end_s = min(start_s + req.clip_length_seconds, video_end_s)

        if end_s - start_s < req.min_clip_seconds:
            continue

        # de-dupe very close start times (ms-level)
        start_ms = int(start_s * 1000)
        if start_ms in used_starts:
            continue
        used_starts.add(start_ms)

        candidates.append(Candidate(
            start_ms=start_ms,
            end_ms=int(end_s * 1000),
            score=1.0,  # placeholder scoring for now
        ))

    # If still empty, fallback to a first 30s chunk
    if not candidates:
        end_s = min(float(req.clip_length_seconds), video_end_s)
        candidates = [Candidate(start_ms=0, end_ms=int(end_s * 1000), score=0.1)]

    return DetectResponse(
        candidates=candidates,
        video_end_seconds=video_end_s,
        scenes_detected=len(scene_list),
    )
