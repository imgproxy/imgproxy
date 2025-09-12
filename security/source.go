package security

// VerifySourceURL checks if the given imageURL is allowed based on
// the configured AllowedSources.
func (s *Security) VerifySourceURL(imageURL string) error {
	if len(s.config.AllowedSources) == 0 {
		return nil
	}

	for _, allowedSource := range s.config.AllowedSources {
		if allowedSource.MatchString(imageURL) {
			return nil
		}
	}

	return newSourceURLError(imageURL)
}
