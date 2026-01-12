def build_merged_clips(event_times: list[float], req, video_end_s: float) -> list[dict]:
    if not event_times:
        return []

    # Convert each event to a window [t-pre, t+post]
    windows = []
    for t in sorted(event_times):
        s = max(0.0, t - float(req.pre_roll_seconds))
        e = min(video_end_s, t + float(req.post_roll_seconds))
        windows.append((s, e))

    # Merge nearby windows using merge_gap_seconds
    merged = []
    cur_s, cur_e = windows[0]
    kill_count = 1
    for s, e in windows[1:]:
        if s <= cur_e + float(req.merge_gap_seconds):
            cur_e = max(cur_e, e)
            kill_count += 1
        else:
            merged.append((cur_s, cur_e, kill_count))
            cur_s, cur_e, kill_count = s, e, 1
    merged.append((cur_s, cur_e, kill_count))

    candidates = []
    for s, e, kcount in merged:
        # enforce min clip length
        if (e - s) < float(req.min_clip_seconds):
            need = float(req.min_clip_seconds) - (e - s)
            s = max(0.0, s - need * 0.5)
            e = min(video_end_s, e + need * 0.5)

        # enforce max clip length by splitting
        max_len = float(req.max_clip_seconds)
        start = s
        while start < e:
            end = min(e, start + max_len)
            if (end - start) < float(req.min_clip_seconds):
                break

            candidates.append({
                "start_ms": int(start * 1000),
                "end_ms": int(end * 1000),
                "score": float(kcount),
            })

            if req.max_candidates > 0 and len(candidates) >= int(req.max_candidates):
                return candidates

            start = end

    return candidates
