package awsconditions

// enum for the results of a condition evaluation
type AWSConditionResult int16

const (
	ConditionTrue       AWSConditionResult = 0
	ConditionFalse      AWSConditionResult = 1
	ConditionUnresolved AWSConditionResult = 2
)

type AWSCondition struct {
	Operator          string
	ConditionKeys     map[string][]string // map of condition keys to values
	Result            AWSConditionResult
	ResolvedVariables map[string]string
}

var functions = map[string]func(string, string) bool{
	"stringequals":              StringEquals,
	"stringnotequals":           StringNotEquals,
	"stringequalsignorecase":    StringEqualsIgnoreCase,
	"stringnotequalsignorecase": StringNotEqualsIgnoreCase,
	"StringLike":                StringLike,
	"StringNotLike":             StringNotLike,
	"DateEquals":                DateEquals,
	"DateNotEquals":             DateNotEquals,
	"DateLessThan":              DateLessThan,
	"DateLessThanEquals":        DateLessThanEquals,
	"DateGreaterThan":           DateGreaterThan,
	"DateGreaterThanEquals":     DateGreaterThanEquals,
	"IpAddress":                 IpAddress,
	"NotIpAddress":              NotIpAddress,
	"ArnEquals":                 ArnEquals,
	"ArnNotEquals":              ArnNotEquals,
	"ArnLike":                   ArnLike,
	"ArnNotLike":                ArnNotLike,
}

func SolveCondition(condition *AWSCondition) bool {

	operatorFunc := functions[condition.Operator]
	if operatorFunc == nil {
		return false
	}

	for conditionKey, conditionValues := range condition.ConditionKeys {
		conditionVal := false
		// condition values are OR'd together
		for _, conditionValue := range conditionValues {
			if operatorFunc(conditionKey, conditionValue) {
				conditionVal = true
				break
			}
		}
		// If none of the values are true, the condition set fails
		if !conditionVal {
			return false
		}
	}
	// If all condition keys are true, the condition set passes
	return true
}
