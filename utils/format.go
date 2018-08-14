package utils


const KB = 1000
const MB = KB * 1000
const GB = MB * 1000
func FormatBytes(b uint64) (float64, string) {
	fb := float64(b)
	if b > GB {
		return fb/ GB, "GB"
	} else if b > MB {
		return fb/ MB, "MB"
	} else if b > KB {
		return fb/ KB, "KB"
	}
	return fb, "B"
}
