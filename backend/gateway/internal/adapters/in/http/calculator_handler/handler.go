package calculatorhandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/victorinolavida/gateway/config"
	authmiddleware "github.com/victorinolavida/gateway/internal/adapters/in/http/middleware/auth_middleware"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/response"
	"github.com/victorinolavida/pkg/jwt"
	calculatorpb "github.com/victorinolavida/proto/calculator"
	"google.golang.org/protobuf/types/known/structpb"
)

type Handler struct {
	client calculatorpb.CalculatorServiceClient
	jwt    *jwt.TokenService
	config *config.Config
}

func New(conf *config.Config, client calculatorpb.CalculatorServiceClient) (*Handler, error) {
	token, err := jwt.NewService(conf.Token)
	if err != nil {
		return nil, err
	}
	return &Handler{client: client, jwt: token, config: conf}, nil
}
func (h *Handler) BasePath() string { return "/calculator" }
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	p := r.With(authmiddleware.AuthMiddleware(h.jwt, &h.config.Cookie))
	p.Post("/price", h.handle("price"))
	p.Post("/yield", h.handle("yield"))
	p.Post("/cashflows", h.handle("cashflows"))
	return r
}

type request struct {
	SecurityID     string `json:"security_id"`
	SettlementDate string `json:"settlement_date"`
	Yield          string `json:"yield"`
	CleanPrice     string `json:"clean_price"`
}

func (h *Handler) handle(method string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := response.ParseBodyFromRequest[request](w, r)
		if err != nil {
			response.RenderBadRequest(w, err.Error())
			return
		}
		if req.SecurityID == "" || (method != "cashflows" && req.SettlementDate == "") || (method == "price" && req.Yield == "") || (method == "yield" && req.CleanPrice == "") {
			response.RenderBadRequest(w, "missing required calculator fields")
			return
		}
		input, err := structpb.NewStruct(map[string]any{"security_id": req.SecurityID, "settlement_date": req.SettlementDate, "yield": req.Yield, "clean_price": req.CleanPrice})
		if err != nil {
			response.RenderBadRequest(w, err.Error())
			return
		}
		var result *structpb.Struct
		switch method {
		case "price":
			result, err = h.client.Price(r.Context(), input)
		case "yield":
			result, err = h.client.Yield(r.Context(), input)
		case "cashflows":
			result, err = h.client.Cashflows(r.Context(), input)
		}
		if err != nil {
			response.RenderGRPCError(w, r, err)
			return
		}
		data := result.AsMap()
		if method == "yield" {
			response.RenderJson(w, data["yield"], http.StatusOK)
			return
		}
		if method == "cashflows" {
			response.RenderJson(w, data["cashflows"], http.StatusOK)
			return
		}
		response.RenderJson(w, data, http.StatusOK)
	}
}
