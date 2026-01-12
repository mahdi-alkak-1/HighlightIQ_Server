# Fortnite Kill Detection Tuning Guide

## Overview

This guide explains how to tune banner-only kill detection to get reliable clips with minimal false positives.

## Recommended Settings (720p)

```json
{
  "max_clip_seconds": 20,
  "pre_roll_seconds": 5,
  "post_roll_seconds": 3,
  "min_clip_seconds": 8,
  "sample_fps": 60,
  "min_spacing_seconds": 1,
  "merge_gap_seconds": 0,
  "elim_match_threshold": 0.6,
  "min_consecutive_hits": 5,
  "cooldown_seconds": 1
}
```

## Parameter Notes

### `elim_match_threshold` (0.0 to 1.0)
- **Purpose**: Template match threshold (TM_CCOEFF_NORMED).
- **Lower** (0.3-0.4): More sensitive, more hits, more false positives.
- **Higher** (0.5-0.7): More strict, fewer hits.
- **If getting 0 hits**: Lower to 0.35-0.4.

### `min_consecutive_hits` (1-60)
- **Purpose**: Require N consecutive frames to confirm a kill.
- **Default**: 3.
- **If missing kills**: Lower to 1-2.
- **If false positives**: Raise to 4-5.

### `cooldown_seconds` (0.0-120.0)
- **Purpose**: Minimum time between detections.
- **If duplicates**: Increase to 1.5-2.0.

### `min_spacing_seconds` (0.0-120.0)
- **Purpose**: Deduplicate events that are too close together.
- **If duplicates**: Increase to 3.0-3.5.

### Clip Length
- `pre_roll_seconds` and `post_roll_seconds` control the capture window around a kill.
- `min_clip_seconds` pads short clips.
- `max_clip_seconds` caps long clips.

## ROI (Fixed for 720p)

ROI is fixed to: x=592, y=535, w=98, h=19. This matches 1280x720 footage.
If you change resolution, update the ROI in code.
