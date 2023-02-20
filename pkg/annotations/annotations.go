package annotations

const (
	mutualCPUsAnnotation = "cpu-mutual.crio.io"
	annotationEnable     = "enable"
)

func IsMutualCPUsEnabled(annot map[string]string) bool {
	return annot[mutualCPUsAnnotation] == annotationEnable
}
