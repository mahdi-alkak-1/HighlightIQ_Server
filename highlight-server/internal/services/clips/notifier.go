package clips

import (
	"context"

	clipsrepo "highlightiq-server/internal/repos/clips"
)

type PublishNotifier interface {
	NotifyClipExported(ctx context.Context, clip clipsrepo.Clip, clipURL string) error
}
