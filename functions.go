package yamlx

import (
	"fmt"
	"github.com/Knetic/govaluate"
	"math/rand"
	"strings"
)

func length(args ...any) (any, error) {
	if strval, ok := args[0].(string); ok {
		length := len(strval)
		return (float64)(length), nil
	} else {
		return len(args), nil
	}
}

func contains(args ...any) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("contains function requires 2 arguments")
	}
	if strval, ok := args[0].(string); ok {
		if substrval, ok := args[1].(string); ok {
			return strings.Contains(strval, substrval), nil
		}
	} else {
		found := false
		for _, v := range args[:len(args)-1] {
			expression, err := govaluate.NewEvaluableExpression(fmt.Sprintf("%v == %v", v, args[len(args)-1]))
			if err != nil {
				return nil, err
			}
			result, err := expression.Evaluate(nil)
			if err != nil {
				return nil, err
			}
			if result.(bool) {
				found = true
				break
			}
		}
		return found, nil
	}
	return nil, fmt.Errorf("contains function requires string arguments")
}

func calcRand(args ...any) (any, error) {
	if min, ok := args[0].(float64); ok && len(args) == 2 {
		if max, ok := args[1].(float64); ok {
			return rand.Int63n(int64(max-min)) + int64(min), nil
		} else {
			return nil, fmt.Errorf("rand function requires 2 integer arguments")
		}
	} else {
		// select random option in args
		return args[rand.Intn(len(args))], nil
	}
	return nil, fmt.Errorf("invalid arguments for rand function")
}

func calcMax(args ...any) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("max function requires at least 1 argument")
	}
	max := args[0]
	for _, v := range args {
		if i, ok := v.(int64); ok {
			if i > max.(int64) {
				max = i
			}
		} else if f, ok := v.(float64); ok {
			if f > max.(float64) {
				max = f
			}
		} else {
			return nil, fmt.Errorf("max function requires numeric arguments")
		}
	}
	return max, nil
}

func calcMin(args ...any) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("min function requires at least 1 argument")
	}
	min := args[0]
	for _, v := range args {
		if i, ok := v.(int64); ok {
			if i < min.(int64) {
				min = i
			}
		} else if f, ok := v.(float64); ok {
			if f < min.(float64) {
				min = f
			}
		} else {
			return nil, fmt.Errorf("min function requires numeric arguments")
		}
	}
	return min, nil
}

func upper(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("upper function requires 1 argument")
	}
	if strval, ok := args[0].(string); ok {
		return strings.ToUpper(strval), nil
	}
	return nil, fmt.Errorf("upper function requires string argument")
}

func lower(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("lower function requires 1 argument")
	}
	if strval, ok := args[0].(string); ok {
		return strings.ToLower(strval), nil
	}
	return nil, fmt.Errorf("lower function requires string argument")
}

func title(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("title function requires 1 argument")
	}
	if strval, ok := args[0].(string); ok {
		return strings.Title(strval), nil
	}
	return nil, fmt.Errorf("title function requires string argument")
}

func trim(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("trim function requires 1 argument")
	}
	if strval, ok := args[0].(string); ok {
		return strings.TrimSpace(strval), nil
	}
	return nil, fmt.Errorf("trim function requires string argument")
}

func join(args ...any) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("join function requires 2 arguments")
	}
	if strval, ok := args[0].(string); ok {
		if slice, ok := args[1].([]any); ok {
			strs := make([]string, len(slice))
			for i, v := range slice {
				strs[i] = fmt.Sprintf("%v", v)
			}
			return strings.Join(strs, strval), nil
		} else {
			// join remaining arguments
			strs := make([]string, len(args)-1)
			for i, v := range args[1:] {
				strs[i] = fmt.Sprintf("%v", v)
			}
			return strings.Join(strs, strval), nil
		}
	}
	return nil, fmt.Errorf("join function requires string and slice arguments")
}

func replace(args ...any) (any, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("replace function requires 3 arguments")
	}
	if strval, ok := args[0].(string); ok {
		if oldstrval, ok := args[1].(string); ok {
			if newstrval, ok := args[2].(string); ok {
				return strings.Replace(strval, oldstrval, newstrval, -1), nil
			}
		}
	}
	return nil, fmt.Errorf("replace function requires string arguments")
}

func substr(args ...any) (any, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("substr function requires 3 arguments")
	}
	if strval, ok := args[0].(string); ok {
		if startval, ok := args[1].(float64); ok {
			if endval, ok := args[2].(float64); ok {
				return strval[int(startval):int(endval)], nil
			}
		}
	}
	return nil, fmt.Errorf("substr function requires string and integer arguments")
}

func strrev(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("strrev function requires 1 argument")
	}
	if strval, ok := args[0].(string); ok {
		runes := []rune(strval)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes), nil
	}
	return nil, fmt.Errorf("strrev function requires string argument")
}

func startswith(args ...any) (any, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("startswith function requires 2 arguments")
	}
	if strval, ok := args[0].(string); ok {
		if substrval, ok := args[1].(string); ok {
			return strings.HasPrefix(strval, substrval), nil
		}
	}
	return nil, fmt.Errorf("startswith function requires string arguments")
}

func endswith(args ...any) (any, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("endswith function requires 2 arguments")
	}
	if strval, ok := args[0].(string); ok {
		if substrval, ok := args[1].(string); ok {
			return strings.HasSuffix(strval, substrval), nil
		}
	}
	return nil, fmt.Errorf("endswith function requires string arguments")
}

func alltrue(args ...any) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("alltrue function requires at least 1 argument")
	}
	for _, v := range args {
		if b, ok := v.(bool); ok {
			if !b {
				return false, nil
			}
		} else {
			return nil, fmt.Errorf("alltrue function requires boolean arguments")
		}
	}
	return true, nil
}

func anytrue(args ...any) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("anytrue function requires at least 1 argument")
	}
	for _, v := range args {
		if b, ok := v.(bool); ok {
			if b {
				return true, nil
			}
		} else {
			return nil, fmt.Errorf("anytrue function requires boolean arguments")
		}
	}
	return false, nil
}
