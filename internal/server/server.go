package server

import (
	"errors"
	"net/http"

	"github.com/10664kls/app-in-performance-api/internal/appin"
	"github.com/labstack/echo/v4"
	edpb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	rpcstatus "google.golang.org/grpc/status"
)

type Server struct {
	appin *appin.Service
}

func NewServer(appin *appin.Service) (*Server, error) {
	if appin == nil {
		return nil, errors.New("appin is nil")
	}

	return &Server{
		appin: appin,
	}, nil
}

func (s *Server) Install(e *echo.Echo, mws ...echo.MiddlewareFunc) error {
	if e == nil {
		return errors.New("echo is nil")
	}

	v1 := e.Group("/v1")

	v1.GET("/appins", s.listAppIns, mws...)
	v1.GET("/appins/overview", s.getAppInOverview, mws...)

	return nil
}

func badJSON() error {
	s, _ := rpcstatus.New(codes.InvalidArgument, "Request body must be a valid JSON.").
		WithDetails(&edpb.ErrorInfo{
			Reason: "BINDING_ERROR",
			Domain: "http",
		})

	return s.Err()
}

func badParam() error {
	s, _ := rpcstatus.New(codes.InvalidArgument, "Request parameters must be a valid type.").
		WithDetails(&edpb.ErrorInfo{
			Reason: "BINDING_ERROR",
			Domain: "http",
		})

	return s.Err()
}

func (s *Server) listAppIns(c echo.Context) error {
	req := new(appin.Query)
	if err := c.Bind(req); err != nil {
		return badParam()
	}

	as, err := s.appin.ListAppIns(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, as)
}

func (s *Server) getAppInOverview(c echo.Context) error {
	req := new(appin.Query)
	if err := c.Bind(req); err != nil {
		return badParam()
	}

	as, err := s.appin.GetOverview(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"overview": as,
	})
}
