# Fortnite Kill Detection - Debug Strategy

## Problem: Getting 0 Clips Detected

`inserted: 0` means no banner events survived the filters.

## Checklist

1. **Templates exist and match the banner**
   - Put all kill-banner templates in `clipper/templates/*.png`.
   - Use 8-10 clean templates from real frames.

2. **ROI is correct for 720p**
   - Fixed ROI: x=592, y=535, w=98, h=19.
   - If ROI is off, template matching will fail.

3. **Threshold too high**
   - Lower `elim_match_threshold` to 0.35-0.4.

4. **Consecutive hits too strict**
   - Lower `min_consecutive_hits` to 1-2 for debugging.

## Minimal Debug Request

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

If this still gives 0 hits, the templates or ROI do not match your video.
