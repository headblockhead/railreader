package main

import (
	"context"
	"errors"
	"net"

	"github.com/huin/goupnp/dcps/internetgateway2"
	"golang.org/x/sync/errgroup"
)

type RouterClient interface {
	AddPortMapping(
		NewRemoteHost string,
		NewExternalPort uint16,
		NewProtocol string,
		NewInternalPort uint16,
		NewInternalClient string,
		NewEnabled bool,
		NewPortMappingDescription string,
		NewLeaseDuration uint32,
	) (err error)

	GetExternalIPAddress() (
		NewExternalIPAddress string,
		err error,
	)

	LocalAddr() net.IP
}

func PickRouterClient(ctx context.Context) (RouterClient, error) {
	searches, _ := errgroup.WithContext(ctx)
	var ip1Clients []*internetgateway2.WANIPConnection1
	searches.Go(func() error {
		var err error
		ip1Clients, _, err = internetgateway2.NewWANIPConnection1ClientsCtx(ctx)
		return err
	})
	var ip2Clients []*internetgateway2.WANIPConnection2
	searches.Go(func() error {
		var err error
		ip2Clients, _, err = internetgateway2.NewWANIPConnection2ClientsCtx(ctx)
		return err
	})
	var ppp1Clients []*internetgateway2.WANPPPConnection1
	searches.Go(func() error {
		var err error
		ppp1Clients, _, err = internetgateway2.NewWANPPPConnection1ClientsCtx(ctx)
		return err
	})

	if err := searches.Wait(); err != nil {
		return nil, err
	}

	if len(ip2Clients) > 0 {
		if len(ip2Clients) > 1 {
			return nil, errors.New("multiple UPnP IGD v2 devices found")
		}
		return ip2Clients[0], nil
	}
	if len(ip1Clients) > 0 {
		if len(ip1Clients) > 1 {
			return nil, errors.New("multiple UPnP IGD v1 devices found")
		}
		return ip1Clients[0], nil
	}
	if len(ppp1Clients) > 0 {
		if len(ppp1Clients) > 1 {
			return nil, errors.New("multiple UPnP PPP devices found")
		}
		return ppp1Clients[0], nil
	}

	return nil, errors.New("no UPnP devices found")
}
