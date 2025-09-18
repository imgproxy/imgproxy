package processing

func applyFilters(с *Context) error {
	if с.PO.Blur == 0 && с.PO.Sharpen == 0 && с.PO.Pixelate <= 1 {
		return nil
	}

	if err := с.Img.CopyMemory(); err != nil {
		return err
	}

	if err := с.Img.RgbColourspace(); err != nil {
		return err
	}

	if err := с.Img.ApplyFilters(с.PO.Blur, с.PO.Sharpen, с.PO.Pixelate); err != nil {
		return err
	}

	return с.Img.CopyMemory()
}
