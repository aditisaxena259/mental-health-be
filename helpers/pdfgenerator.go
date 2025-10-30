package helpers

import (
	"fmt"
	"github.com/aditisaxena259/mental-health-be/models"
	"github.com/jung-kurt/gofpdf"
	"time"
)

func CreatePDF(apologies []models.Apology) (string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(40, 10, "Monthly Apology Report")

	pdf.Ln(12)
	pdf.SetFont("Arial", "", 10)

	for _, a := range apologies {
		pdf.Cell(190, 10, fmt.Sprintf("StudentID: %s | Type: %s | Status: %s | Date: %s", a.StudentID, a.ApologyType, a.Status, a.CreatedAt.Format("2006-01-02")))
		pdf.Ln(8)
		pdf.MultiCell(190, 6, fmt.Sprintf("Comment: %s\nDescription: %s", a.Comment, a.Description), "", "", false)
		pdf.Ln(4)
	}

	filename := fmt.Sprintf("apology_report_%d.pdf", time.Now().Unix())
	path := fmt.Sprintf("./reports/%s", filename)
	err := pdf.OutputFileAndClose(path)

	return path, err
}
