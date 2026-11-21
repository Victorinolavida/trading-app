package transport

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/victorinolavida/calculator/api"
	calculatorpb "github.com/victorinolavida/proto/calculator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/structpb"
	"net"
	"testing"
	"time"
)

type stub struct{ security api.Security }

func (s stub) GetByPublicID(context.Context, uuid.UUID) (api.Security, error) { return s.security, nil }
func TestGRPCPriceAndCashflows(t *testing.T) {
	id := uuid.New()
	settlement := time.Date(2027, 1, 5, 0, 0, 0, 0, time.UTC)
	nominal, err := api.ParseDecimal("100")
	require.NoError(t, err)
	module, err := api.New(api.Deps{Security: stub{api.NewBondSecurity(id.String(), "", "M 5", "M", "MXN", settlement.AddDate(1, 0, 0), nominal)}})
	require.NoError(t, err)
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(server, &Server{Module: module})
	go server.Serve(listener)
	t.Cleanup(server.Stop)
	conn, err := grpc.NewClient("passthrough:///calculator", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }))
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })
	client := calculatorpb.NewCalculatorServiceClient(conn)
	request, err := structpb.NewStruct(map[string]any{"security_id": id.String(), "settlement_date": settlement.Format(time.RFC3339), "yield": "0.05"})
	require.NoError(t, err)
	result, err := client.Price(t.Context(), request)
	require.NoError(t, err)
	require.NotEmpty(t, result.Fields["clean_price"].GetStringValue())
	flows, err := client.Cashflows(t.Context(), request)
	require.NoError(t, err)
	require.NotEmpty(t, flows.Fields["cashflows"].GetListValue().Values)
}
