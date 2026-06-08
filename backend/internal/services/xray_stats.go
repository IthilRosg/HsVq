package services

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/xtls/xray-core/app/stats/command"
)

const xrayGrpcAddr = "127.0.0.1:62789"

type XrayStatsClient struct {
	conn *grpc.ClientConn
}

func NewXrayStatsClient() (*XrayStatsClient, error) {
	conn, err := grpc.Dial(xrayGrpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to xray gRPC: %v", err)
	}
	return &XrayStatsClient{conn: conn}, nil
}

func (c *XrayStatsClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// GetUserTraffic возвращает трафик пользователя по inbound email
func (c *XrayStatsClient) GetUserTraffic(email string, reset bool) (up int64, down int64, err error) {
	client := command.NewStatsServiceClient(c.conn)

	for _, name := range []string{
		fmt.Sprintf("user>>>%s>>>traffic>>>uplink", email),
		fmt.Sprintf("user>>>%s>>>traffic>>>downlink", email),
	} {
		req := &command.GetStatsRequest{
			Name:   name,
			Reset_: reset,
		}
		resp, err := client.GetStats(context.Background(), req)
		if err != nil {
			continue
		}
		if resp.Stat != nil {
			if name == fmt.Sprintf("user>>>%s>>>traffic>>>uplink", email) {
				up = resp.Stat.Value
			} else {
				down = resp.Stat.Value
			}
		}
	}
	return
}

// GetAllTraffic возвращает трафик всех пользователей
func (c *XrayStatsClient) GetAllTraffic() (map[string][2]int64, error) {
	client := command.NewStatsServiceClient(c.conn)
	result := make(map[string][2]int64)

	req := &command.QueryStatsRequest{
		Pattern: "user>>>",
		Reset_:  false,
	}
	resp, err := client.QueryStats(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("query stats: %v", err)
	}

	for _, stat := range resp.Stat {
		// Парсим email из имени "user>>>email>>>traffic>>>uplink"
		var email string
		var isUp bool
		_, _ = fmt.Sscanf(stat.Name, "user>>>%s>>>traffic>>>uplink", &email)
		if email != "" {
			isUp = true
		} else {
			_, _ = fmt.Sscanf(stat.Name, "user>>>%s>>>traffic>>>downlink", &email)
		}
		if email == "" {
			continue
		}
		entry := result[email]
		if isUp {
			entry[0] = stat.Value
		} else {
			entry[1] = stat.Value
		}
		result[email] = entry
	}
	return result, nil
}

// GetServerTraffic возвращает общий трафик сервера (inbound)
func (c *XrayStatsClient) GetServerTraffic() (up int64, down int64, err error) {
	client := command.NewStatsServiceClient(c.conn)

	for _, name := range []string{
		"inbound>>>reality-in>>>traffic>>>uplink",
		"inbound>>>reality-in>>>traffic>>>downlink",
	} {
		req := &command.GetStatsRequest{Name: name, Reset_: false}
		resp, err := client.GetStats(context.Background(), req)
		if err != nil {
			continue
		}
		if resp.Stat != nil {
			if name == "inbound>>>reality-in>>>traffic>>>uplink" {
				up = resp.Stat.Value
			} else {
				down = resp.Stat.Value
			}
		}
	}
	return
}
