package check

import (
	"github.com/honestbank/tech-assignment-backend-engineer/db"
	"net/http"
	"strconv"

	. "github.com/honestbank/tech-assignment-backend-engineer/constants"
	"github.com/honestbank/tech-assignment-backend-engineer/model"
	"github.com/honestbank/tech-assignment-backend-engineer/writer"
)

var Writer writer.IWriter = &writer.WriterImpl{}

type NumberPreApprovedCheck struct {
	next ICheck
}

type AgeCheck struct {
	next ICheck
}

type AreaCodeCheck struct {
	next ICheck
}

type IncomeCheck struct {
	next ICheck
}

type NumberOfCreditCardsCheck struct {
	next ICheck
}

type CreditRiskScoreCheck struct {
	next ICheck
}

type PoliticallyExposedCheck struct {
	next ICheck
}

func (n *NumberPreApprovedCheck) Check(data model.RecordData, uid string) (bool, int, error) {
	flag, err := db.CheckIfNumberPresent(data.PhoneNumber)
	if err != nil {
		return false, http.StatusServiceUnavailable, err
	}
	return flag, http.StatusOK, err
}

func (a *AgeCheck) Check(data model.RecordData, uid string) (bool, int, error) {
	if data.Age >= MIN_AGE {
		if a.next != nil {
			return a.next.Check(data, uid)
		}
		return true, http.StatusOK, nil
	}
	Writer.LogToJSON(uid, INVALID_AGE, DECLINED, LOG_LEVEL_WARN)
	return false, http.StatusOK, nil
}

func (a *AreaCodeCheck) Check(data model.RecordData, uid string) (bool, int, error) {
	areaCodeStr := string(data.PhoneNumber[0])
	areaCode, err := strconv.Atoi(areaCodeStr)
	if err != nil {
		Writer.LogToJSON(uid, INVALID_AREA_CODE, DECLINED, LOG_LEVEL_ERROR)
		return false, http.StatusInternalServerError, err
	}
	for _, code := range ALLOWED_AREA_CODE {
		if areaCode == code {
			if a.next != nil {
				return a.next.Check(data, uid)
			}
			return true, http.StatusOK, nil
		}
	}
	Writer.LogToJSON(uid, INVALID_AREA_CODE, DECLINED, LOG_LEVEL_WARN)
	return false, http.StatusOK, nil
}

func (n *NumberOfCreditCardsCheck) Check(data model.RecordData, uid string) (bool, int, error) {
	if data.NumberOfCreditCards != nil && *data.NumberOfCreditCards <= MAX_NUMBER_OF_CC {
		if n.next != nil {
			return n.next.Check(data, uid)
		}
		return true, http.StatusOK, nil
	}
	Writer.LogToJSON(uid, INVALID_CC_NUMBER, DECLINED, LOG_LEVEL_WARN)
	return false, http.StatusOK, nil
}

func (i *IncomeCheck) Check(data model.RecordData, uid string) (bool, int, error) {
	if data.Income >= MIN_INCOME {
		if i.next != nil {
			return i.next.Check(data, uid)
		}
		return true, http.StatusOK, nil
	}
	Writer.LogToJSON(uid, INVALID_INCOME, DECLINED, LOG_LEVEL_WARN)
	return false, http.StatusOK, nil
}

func (c *CreditRiskScoreCheck) Check(data model.RecordData, uid string) (bool, int, error) {
	if data.NumberOfCreditCards != nil &&
		DESIRED_CREDIT_RISK_SCORE == calculateCreditRisk(data.Age, *data.NumberOfCreditCards) {
		if c.next != nil {
			return c.next.Check(data, uid)
		}
		return true, http.StatusOK, nil
	}
	Writer.LogToJSON(uid, INVALID_CREDIT_RISK_SCORE, DECLINED, LOG_LEVEL_WARN)
	return false, http.StatusOK, nil
}

func (p *PoliticallyExposedCheck) Check(data model.RecordData, uid string) (bool, int, error) {
	if data.PoliticallyExposed != nil && *data.PoliticallyExposed {
		Writer.LogToJSON(uid, POLITICALLY_EXPOSED, DECLINED, LOG_LEVEL_WARN)
		return false, http.StatusOK, nil
	}
	if p.next != nil {
		return p.next.Check(data, uid)
	}
	return true, http.StatusOK, nil
}

// CreateChecks creates the instances of all checks and sets up the chain of responsibility.
// It returns the first check in the chain
func CreateEligibilityChecks() ICheck {
	// instance creation
	ageCheck := &AgeCheck{}
	areaCodeCheck := &AreaCodeCheck{}
	incomeCheck := &IncomeCheck{}
	numberOfCreditCardsCheck := &NumberOfCreditCardsCheck{}
	creditRiskScoreCheck := &CreditRiskScoreCheck{}
	politicallyExposedCheck := &PoliticallyExposedCheck{}

	// Set up the chain
	ageCheck.SetNext(areaCodeCheck)
	areaCodeCheck.SetNext(incomeCheck)
	incomeCheck.SetNext(numberOfCreditCardsCheck)
	numberOfCreditCardsCheck.SetNext(creditRiskScoreCheck)
	creditRiskScoreCheck.SetNext(politicallyExposedCheck)

	// Return the first check in the chain
	return ageCheck
}

func CreatePhoneNumberCheck() ICheck {
	return &NumberPreApprovedCheck{}
}

// calculateCreditRisk calculates the credit risk score based on the age and number of credit cards.
func calculateCreditRisk(age, numberOfCreditCard int) string {
	sum := age + numberOfCreditCard
	mod := sum % 3
	if mod == 0 {
		return "LOW"
	}
	if mod == 1 {
		return "MEDIUM"
	}
	return "HIGH"
}

func (n *NumberPreApprovedCheck) SetNext(check ICheck) {
	n.next = check
}
func (a *AgeCheck) SetNext(check ICheck) {
	a.next = check
}
func (a *AreaCodeCheck) SetNext(check ICheck) {
	a.next = check
}
func (n *NumberOfCreditCardsCheck) SetNext(check ICheck) {
	n.next = check
}
func (i *IncomeCheck) SetNext(check ICheck) {
	i.next = check
}
func (c *CreditRiskScoreCheck) SetNext(check ICheck) {
	c.next = check
}
func (p *PoliticallyExposedCheck) SetNext(check ICheck) {
	p.next = check
}
