package clips

type ExportRequest struct {
	// keep empty for now; later we can add format/preset/etc.
}

func (r ExportRequest) Validate() error {
	return nil
}
