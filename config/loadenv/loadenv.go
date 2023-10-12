package loadenv

func Load() error {
	if err := loadAWSSecret(); err != nil {
		return err
	}

	if err := loadAWSSystemManagerParams(); err != nil {
		return err
	}

	if err := loadGCPSecret(); err != nil {
		return err
	}

	if err := loadLocalFile(); err != nil {
		return err
	}

	return nil
}
