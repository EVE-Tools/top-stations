package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/EVE-Tools/element43/go/lib/transport"
	"github.com/EVE-Tools/top-stations/lib/esiMarkets"
	pb "github.com/EVE-Tools/top-stations/lib/topStations"
	"github.com/antihax/goesi"
	"github.com/boltdb/bolt"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/kelseyhightower/envconfig"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Config holds the application's configuration info from the environment.
type Config struct {
	Cron           string `default:"@hourly" envconfig:"cron"`
	DBPath         string `default:"top-stations.db" envconfig:"db_path"`
	LogLevel       string `default:"info" envconfig:"log_level"`
	EsiMarketsHost string `default:"esi-markets.element43.svc.cluster.local:43000" envconfig:"esi_markets_host"`
	Port           string `default:"43000" envconfig:"port"`
}

// Server is a gRPC server serving location requests
type Server struct {
	db *bolt.DB
}

// Global instances, move into context/parameter if project grows larger!
var db *bolt.DB
var config *Config

// GetTopStations returns the top stations from persistence.
func (server *Server) GetTopStations(context context.Context, request *pb.GetTopStationsRequest) (*pb.GetTopStationsResponse, error) {
	var statsBlob []byte

	// Try to get stats from BoltDB
	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("topStations"))
		statsBlob = bucket.Get([]byte("stats"))
		return nil
	})

	if statsBlob == nil {
		logrus.Error("could not get stats from BoltDB")
		return nil, status.Error(codes.NotFound, "Error retrieving stats")
	}

	var stats pb.GetTopStationsResponse
	err := proto.Unmarshal(statsBlob, &stats)
	if err != nil {
		logrus.WithError(err).Error("could not parse stats from BoltDB")
		return nil, status.Error(codes.NotFound, "Error parsing stats")
	}

	return &stats, nil
}

func main() {
	loadConfig()
	startEndpoint()

	// Terminate this goroutine, crash if all other goroutines exited
	runtime.Goexit()
}

// Load configuration from environment
func loadConfig() {
	config = &Config{}
	envconfig.MustProcess("TOP_STATIONS", config)

	logLevel, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		panic(err)
	}

	logrus.SetLevel(logLevel)
	logrus.Debugf("Config: %+v", config)
}

// Init DB and start gRPC endpoint.
func startEndpoint() {
	var err error
	db, err = bolt.Open(config.DBPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		panic(err)
	}

	// Initialize buckets
	err = db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("topStations"))
		return nil
	})
	if err != nil {
		panic(err)
	}

	// Start cron
	job := cron.New()
	job.AddFunc(config.Cron, updateTopStations)
	job.Start()

	// Inititalize gRPC server
	var opts []grpc.ServerOption
	var logOpts []grpc_logrus.Option

	opts = append(opts, grpc_middleware.WithUnaryServerChain(
		grpc_ctxtags.UnaryServerInterceptor(),
		grpc_logrus.UnaryServerInterceptor(logrus.NewEntry(logrus.New()), logOpts...)))

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", config.Port))

	if err != nil {
		log.Fatalf("could not listen: %v", err)
	}

	updateTopStations()

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterTopStationsServer(grpcServer, &Server{db: db})
	grpcServer.Serve(listener)
}

func updateTopStations() {
	//
	// 1. Download list of region_ids
	// 2. For each region download orders and aggregate stats
	// 3. Save to BoltDB
	//

	const userAgent string = "Element43/top-stations (element-43.com)"
	const timeout time.Duration = time.Duration(time.Second * 30)

	httpClientESI := &http.Client{
		Timeout:   timeout,
		Transport: transport.NewESITransport(userAgent, timeout),
	}

	esiClient := goesi.NewAPIClient(httpClientESI, userAgent)

	regionIDs, _, err := esiClient.ESI.UniverseApi.GetUniverseRegions(nil, nil)

	if err != nil {
		logrus.WithError(err).Fatal("could not load regions from ESI")
	}

	// Set up a connection to the server.
	opts := []grpc.DialOption{grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(50000000)), grpc.WithInsecure()}
	conn, err := grpc.Dial(config.EsiMarketsHost, opts...)
	if err != nil {
		logrus.WithError(err).Fatal("could not connect to esiMarkets")
	}
	defer conn.Close()

	esiMarketsClient := esiMarkets.NewESIMarketsClient(conn)

	statsMap := make(map[uint64]*pb.StationStats)

	for _, id := range regionIDs {
		logrus.Infof("Processing region %d...", id)

		orders, err := esiMarketsClient.GetRegion(context.Background(), &esiMarkets.GetRegionRequest{RegionId: uint64(id)})
		if err != nil {
			logrus.WithError(err).WithField("region_id", id).Warn("could not load orders from esiMarkets for region")
			continue
		}

		for _, order := range orders.Orders {
			if stats, ok := statsMap[order.LocationId]; ok {
				stats.TotalOrders = stats.TotalOrders + 1
				stats.TotalVolume = stats.TotalVolume + (order.Price * float64(order.VolumeRemain))

				if order.IsBuyOrder {
					stats.BidVolume = stats.BidVolume + (order.Price * float64(order.VolumeRemain))
				} else {
					stats.AskVolume = stats.AskVolume + (order.Price * float64(order.VolumeRemain))
				}
			} else {
				if order.IsBuyOrder {
					statsMap[order.LocationId] = &pb.StationStats{
						Id:          int64(order.LocationId),
						AskVolume:   0,
						BidVolume:   order.Price * float64(order.VolumeRemain),
						TotalVolume: order.Price * float64(order.VolumeRemain),
						TotalOrders: 1,
					}
				} else {
					statsMap[order.LocationId] = &pb.StationStats{
						Id:          int64(order.LocationId),
						AskVolume:   order.Price * float64(order.VolumeRemain),
						BidVolume:   0,
						TotalVolume: order.Price * float64(order.VolumeRemain),
						TotalOrders: 1,
					}
				}
			}
		}
	}

	stats := pb.GetTopStationsResponse{
		Stations: make([]*pb.StationStats, 0),
	}

	for _, value := range statsMap {
		stats.Stations = append(stats.Stations, value)
	}

	blob, err := proto.Marshal(&stats)
	if err != nil {
		logrus.WithError(err).Warn("could not marshal station stats")
		return
	}

	db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("topStations"))
		if bucket == nil {
			panic("Bucket not found! This should never happen!")
		}

		err := bucket.Put([]byte("stats"), blob)
		return err
	})

	logrus.Info("Done!")
}
