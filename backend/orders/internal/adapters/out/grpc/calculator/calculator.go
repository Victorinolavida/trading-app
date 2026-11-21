package calculator

import (
	"context"
	"fmt"
	"strconv"

	calculatorpb "github.com/victorinolavida/proto/calculator"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

type Client struct {
	client calculatorpb.CalculatorServiceClient
}

func New(conn grpc.ClientConnInterface) *Client {
	return &Client{client: calculatorpb.NewCalculatorServiceClient(conn)}
}
func (c *Client) Calculate(ctx context.Context, securityID, settlement, field, value string) (string, error) {
	req, err := structpb.NewStruct(map[string]any{"security_id": securityID, "settlement_date": settlement, field: value})
	if err != nil {
		return "", err
	}
	var res *structpb.Struct
	var key string
	if field == "yield" {
		res, err = c.client.Price(ctx, req)
		key = "clean_price"
	} else {
		res, err = c.client.Yield(ctx, req)
		key = "yield"
	}
	if err != nil {
		return "", err
	}
	v, ok := res.GetFields()[key]
	if !ok {
		return "", fmt.Errorf("calculator response missing %s", key)
	}
	switch x := v.GetKind().(type) {
	case *structpb.Value_StringValue:
		return x.StringValue, nil
	case *structpb.Value_NumberValue:
		return strconv.FormatFloat(x.NumberValue, 'f', -1, 64), nil
	default:
		return "", fmt.Errorf("calculator response has invalid %s", key)
	}
}
