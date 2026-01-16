package youtubepublishes

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

var ErrNotFound = errors.New("youtube_publishes: not found")

type Repo struct {
	db *sql.DB
}

func New(db *sql.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, p CreateParams) (YoutubePublish, error) {
	if p.Status == "" {
		p.Status = "uploaded"
	}

	const q = `
		INSERT INTO youtube_publishes (
			clip_id, youtube_video_id, youtube_url, status, published_at, last_synced_at, views, likes, comments, analytics
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	res, err := r.db.ExecContext(ctx, q,
		p.ClipID,
		p.YoutubeVideoID,
		p.YoutubeURL,
		p.Status,
		p.PublishedAt,
		p.LastSyncedAt,
		p.Views,
		p.Likes,
		p.Comments,
		p.Analytics,
	)
	if err != nil {
		return YoutubePublish{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return YoutubePublish{}, err
	}

	return r.GetByID(ctx, id)
}

func (r *Repo) GetByID(ctx context.Context, id int64) (YoutubePublish, error) {
	const q = `
		SELECT id, clip_id, youtube_video_id, youtube_url, status, published_at, last_synced_at,
		       views, likes, comments, analytics, created_at, updated_at
		FROM youtube_publishes
		WHERE id = ?
		LIMIT 1
	`

	var yp YoutubePublish
	var publishedAt sql.NullTime
	var lastSyncedAt sql.NullTime
	var analytics sql.NullString

	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&yp.ID,
		&yp.ClipID,
		&yp.YoutubeVideoID,
		&yp.YoutubeURL,
		&yp.Status,
		&publishedAt,
		&lastSyncedAt,
		&yp.Views,
		&yp.Likes,
		&yp.Comments,
		&analytics,
		&yp.CreatedAt,
		&yp.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return YoutubePublish{}, ErrNotFound
	}
	if err != nil {
		return YoutubePublish{}, err
	}

	if publishedAt.Valid {
		t := publishedAt.Time
		yp.PublishedAt = &t
	}
	if lastSyncedAt.Valid {
		t := lastSyncedAt.Time
		yp.LastSyncedAt = &t
	}
	if analytics.Valid {
		v := analytics.String
		yp.Analytics = &v
	}

	return yp, nil
}

func (r *Repo) GetByIDForUser(ctx context.Context, userID int64, id int64) (YoutubePublish, error) {
	const q = `
		SELECT yp.id, yp.clip_id, yp.youtube_video_id, yp.youtube_url, yp.status, yp.published_at, yp.last_synced_at,
		       yp.views, yp.likes, yp.comments, yp.analytics, yp.created_at, yp.updated_at
		FROM youtube_publishes yp
		JOIN clips c ON c.id = yp.clip_id
		WHERE c.user_id = ? AND yp.id = ?
		LIMIT 1
	`

	var yp YoutubePublish
	var publishedAt sql.NullTime
	var lastSyncedAt sql.NullTime
	var analytics sql.NullString

	err := r.db.QueryRowContext(ctx, q, userID, id).Scan(
		&yp.ID,
		&yp.ClipID,
		&yp.YoutubeVideoID,
		&yp.YoutubeURL,
		&yp.Status,
		&publishedAt,
		&lastSyncedAt,
		&yp.Views,
		&yp.Likes,
		&yp.Comments,
		&analytics,
		&yp.CreatedAt,
		&yp.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return YoutubePublish{}, ErrNotFound
	}
	if err != nil {
		return YoutubePublish{}, err
	}

	if publishedAt.Valid {
		t := publishedAt.Time
		yp.PublishedAt = &t
	}
	if lastSyncedAt.Valid {
		t := lastSyncedAt.Time
		yp.LastSyncedAt = &t
	}
	if analytics.Valid {
		v := analytics.String
		yp.Analytics = &v
	}

	return yp, nil
}

func (r *Repo) GetByVideoID(ctx context.Context, youtubeVideoID string) (YoutubePublish, error) {
	const q = `
		SELECT id, clip_id, youtube_video_id, youtube_url, status, published_at, last_synced_at,
		       views, likes, comments, analytics, created_at, updated_at
		FROM youtube_publishes
		WHERE youtube_video_id = ?
		LIMIT 1
	`

	var yp YoutubePublish
	var publishedAt sql.NullTime
	var lastSyncedAt sql.NullTime
	var analytics sql.NullString

	err := r.db.QueryRowContext(ctx, q, youtubeVideoID).Scan(
		&yp.ID,
		&yp.ClipID,
		&yp.YoutubeVideoID,
		&yp.YoutubeURL,
		&yp.Status,
		&publishedAt,
		&lastSyncedAt,
		&yp.Views,
		&yp.Likes,
		&yp.Comments,
		&analytics,
		&yp.CreatedAt,
		&yp.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return YoutubePublish{}, ErrNotFound
	}
	if err != nil {
		return YoutubePublish{}, err
	}

	if publishedAt.Valid {
		t := publishedAt.Time
		yp.PublishedAt = &t
	}
	if lastSyncedAt.Valid {
		t := lastSyncedAt.Time
		yp.LastSyncedAt = &t
	}
	if analytics.Valid {
		v := analytics.String
		yp.Analytics = &v
	}

	return yp, nil
}

func (r *Repo) ListByClipIDForUser(ctx context.Context, userID int64, clipID int64) ([]YoutubePublish, error) {
	const q = `
		SELECT yp.id, yp.clip_id, yp.youtube_video_id, yp.youtube_url, yp.status, yp.published_at, yp.last_synced_at,
		       yp.views, yp.likes, yp.comments, yp.analytics, yp.created_at, yp.updated_at
		FROM youtube_publishes yp
		JOIN clips c ON c.id = yp.clip_id
		WHERE c.user_id = ? AND yp.clip_id = ?
		ORDER BY yp.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, q, userID, clipID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []YoutubePublish
	for rows.Next() {
		var yp YoutubePublish
		var publishedAt sql.NullTime
		var lastSyncedAt sql.NullTime
		var analytics sql.NullString

		if err := rows.Scan(
			&yp.ID,
			&yp.ClipID,
			&yp.YoutubeVideoID,
			&yp.YoutubeURL,
			&yp.Status,
			&publishedAt,
			&lastSyncedAt,
			&yp.Views,
			&yp.Likes,
			&yp.Comments,
			&analytics,
			&yp.CreatedAt,
			&yp.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if publishedAt.Valid {
			t := publishedAt.Time
			yp.PublishedAt = &t
		}
		if lastSyncedAt.Valid {
			t := lastSyncedAt.Time
			yp.LastSyncedAt = &t
		}
		if analytics.Valid {
			v := analytics.String
			yp.Analytics = &v
		}

		out = append(out, yp)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *Repo) ListVideoIDs(ctx context.Context) ([]string, error) {
	const q = `
		SELECT DISTINCT youtube_video_id
		FROM youtube_publishes
		ORDER BY youtube_video_id
	`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *Repo) UpdateByIDForUser(ctx context.Context, userID int64, id int64, p UpdateParams) (YoutubePublish, error) {
	// Ensure the record belongs to the user.
	if _, err := r.GetByIDForUser(ctx, userID, id); err != nil {
		return YoutubePublish{}, err
	}

	setParts := make([]string, 0, 8)
	args := make([]interface{}, 0, 10)

	if p.YoutubeURL != nil {
		setParts = append(setParts, "youtube_url = ?")
		args = append(args, *p.YoutubeURL)
	}
	if p.Status != nil {
		setParts = append(setParts, "status = ?")
		args = append(args, *p.Status)
	}
	if p.PublishedAt != nil {
		setParts = append(setParts, "published_at = ?")
		args = append(args, *p.PublishedAt)
	}
	if p.LastSyncedAt != nil {
		setParts = append(setParts, "last_synced_at = ?")
		args = append(args, *p.LastSyncedAt)
	}
	if p.Views != nil {
		setParts = append(setParts, "views = ?")
		args = append(args, *p.Views)
	}
	if p.Likes != nil {
		setParts = append(setParts, "likes = ?")
		args = append(args, *p.Likes)
	}
	if p.Comments != nil {
		setParts = append(setParts, "comments = ?")
		args = append(args, *p.Comments)
	}
	if p.Analytics != nil {
		setParts = append(setParts, "analytics = ?")
		args = append(args, *p.Analytics)
	}

	if len(setParts) == 0 {
		return r.GetByIDForUser(ctx, userID, id)
	}

	q := `
		UPDATE youtube_publishes
		SET ` + strings.Join(setParts, ", ") + `
		WHERE id = ?
		LIMIT 1
	`

	args = append(args, id)

	res, err := r.db.ExecContext(ctx, q, args...)
	if err != nil {
		return YoutubePublish{}, err
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return YoutubePublish{}, err
	}
	if aff == 0 {
		return YoutubePublish{}, ErrNotFound
	}

	return r.GetByIDForUser(ctx, userID, id)
}

func (r *Repo) UpdateByVideoID(ctx context.Context, youtubeVideoID string, p UpdateParams) (YoutubePublish, error) {
	if _, err := r.GetByVideoID(ctx, youtubeVideoID); err != nil {
		return YoutubePublish{}, err
	}

	setParts := make([]string, 0, 8)
	args := make([]interface{}, 0, 10)

	if p.YoutubeURL != nil {
		setParts = append(setParts, "youtube_url = ?")
		args = append(args, *p.YoutubeURL)
	}
	if p.Status != nil {
		setParts = append(setParts, "status = ?")
		args = append(args, *p.Status)
	}
	if p.PublishedAt != nil {
		setParts = append(setParts, "published_at = ?")
		args = append(args, *p.PublishedAt)
	}
	if p.LastSyncedAt != nil {
		setParts = append(setParts, "last_synced_at = ?")
		args = append(args, *p.LastSyncedAt)
	}
	if p.Views != nil {
		setParts = append(setParts, "views = ?")
		args = append(args, *p.Views)
	}
	if p.Likes != nil {
		setParts = append(setParts, "likes = ?")
		args = append(args, *p.Likes)
	}
	if p.Comments != nil {
		setParts = append(setParts, "comments = ?")
		args = append(args, *p.Comments)
	}
	if p.Analytics != nil {
		setParts = append(setParts, "analytics = ?")
		args = append(args, *p.Analytics)
	}

	if len(setParts) == 0 {
		return r.GetByVideoID(ctx, youtubeVideoID)
	}

	q := `
		UPDATE youtube_publishes
		SET ` + strings.Join(setParts, ", ") + `
		WHERE youtube_video_id = ?
		LIMIT 1
	`

	args = append(args, youtubeVideoID)

	res, err := r.db.ExecContext(ctx, q, args...)
	if err != nil {
		return YoutubePublish{}, err
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return YoutubePublish{}, err
	}
	if aff == 0 {
		return YoutubePublish{}, ErrNotFound
	}

	return r.GetByVideoID(ctx, youtubeVideoID)
}
