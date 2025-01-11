package answerbox

import "github.com/PuerkitoBio/goquery"

type MathBoxContent struct {
	Expression string
	Result     string
}

func ExtractMathBox(doc *goquery.Document) *AnswerBox {
	if mathBox := doc.Find("div.card-section"); mathBox.Length() > 0 {
		mathContent := &MathBoxContent{}

		// Expression
		if expression := mathBox.Find("div.tyYmIf div.BRpYC div.XH1CIc span.vUGUtc"); expression.Length() > 0 {
			mathContent.Expression = expression.Text()
		}

		// Result
		if result := mathBox.Find("div.tyYmIf div.fB3vD div.jlkklc div.z7BZJb span.qv3Wpe"); result.Length() > 0 {
			mathContent.Result = result.Text()
		}

		answerBox := &AnswerBox{
			Type:    "math",
			Content: mathContent,
		}

		return answerBox
	}
	return nil
}
