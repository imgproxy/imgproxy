package processing

import (
	"bytes"

	"github.com/trimmer-io/go-xmp/xmp"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagemeta/iptc"
	"github.com/imgproxy/imgproxy/v3/imagemeta/photoshop"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func stripPS3(img *vips.Image) []byte {
	ps3Data, err := img.GetBlob("iptc-data")
	if err != nil || len(ps3Data) == 0 {
		return nil
	}

	ps3Map := make(photoshop.PhotoshopMap)
	photoshop.Parse(ps3Data, ps3Map)

	iptcData, found := ps3Map[photoshop.IptcKey]
	if !found {
		return nil
	}

	iptcMap := make(iptc.IptcMap)
	err = iptc.Parse(iptcData, iptcMap)
	if err != nil {
		return nil
	}

	for key := range iptcMap {
		if key.RecordID == 2 && key.TagID != 80 && key.TagID != 110 && key.TagID != 116 {
			delete(iptcMap, key)
		}
	}

	if len(iptcMap) == 0 {
		return nil
	}

	ps3Map = photoshop.PhotoshopMap{
		photoshop.IptcKey: iptcMap.Dump(),
	}

	return ps3Map.Dump()
}

func stripXMP(img *vips.Image) []byte {
	xmpData, err := img.GetBlob("xmp-data")
	if err != nil || len(xmpData) == 0 {
		return nil
	}

	xmpDoc, err := xmp.Read(bytes.NewReader(xmpData))
	if err != nil {
		return nil
	}

	namespaces := xmpDoc.Namespaces()
	filteredNs := namespaces[:0]

	for _, ns := range namespaces {
		if ns.Name == "dc" || ns.Name == "xmpRights" || ns.Name == "cc" {
			filteredNs = append(filteredNs, ns)
		}
	}
	xmpDoc.FilterNamespaces(filteredNs)

	nodes := xmpDoc.Nodes()
	for _, n := range nodes {
		if n.Name() == "dc" {
			filteredNodes := n.Nodes[:0]
			for _, nn := range n.Nodes {
				name := nn.Name()
				if name == "rights" || name == "contributor" || name == "creator" || name == "publisher" {
					filteredNodes = append(filteredNodes, nn)
				}
			}
			n.Nodes = filteredNodes

			filteredAttrs := n.Attr[:0]
			for _, a := range n.Attr {
				name := a.Name.Local
				if name == "dc:rights" || name == "dc:contributor" || name == "dc:creator" || name == "dc:publisher" {
					filteredAttrs = append(filteredAttrs, a)
				}
			}
			n.Attr = filteredAttrs
		}
	}

	if len(xmpDoc.Nodes()) == 0 {
		return nil
	}

	xmpData, err = xmp.Marshal(xmpDoc)
	if err != nil {
		return nil
	}

	return xmpData
}

func stripMetadata(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	if !po.StripMetadata {
		return nil
	}

	var ps3Data, xmpData []byte

	if po.KeepCopyright {
		ps3Data = stripPS3(img)
		xmpData = stripXMP(img)
	}

	if err := img.Strip(po.KeepCopyright); err != nil {
		return err
	}

	if po.KeepCopyright {
		if len(ps3Data) > 0 {
			img.SetBlob("iptc-data", ps3Data)
		}

		if len(xmpData) > 0 {
			img.SetBlob("xmp-data", xmpData)
		}
	}

	return nil
}
