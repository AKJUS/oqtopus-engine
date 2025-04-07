package mitig

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/oqtopus-team/oqtopus-engine/coreapp/core"
	pb "github.com/oqtopus-team/oqtopus-engine/coreapp/mitig/mitigation_interface/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TODO: use setting
var mitigator_port = "5011"

type MitigationInfo struct {
	NeedToBeMitigated bool
	Mitigated         bool
	
	Readout string
}

func PseudoInverseMitigation(jd *core.JobData) {
	numOfQubits, err := getNumOfQubits(jd.Result.Counts)
	if err != nil {
		zap.L().Error("failed to get number of qubits/reason: ", zap.Error(err))
		jd.Status = core.FAILED
		return
	}

	// TODO: Allow to be set by parameters
	host := "localhost"
	opts := grpc.WithTransportCredentials(insecure.NewCredentials())

	// connect server
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", host, mitigator_port), opts)
	if err != nil {
		zap.L().Error(fmt.Sprintf("did not connect: %v", err))
		return
	}
	defer conn.Close()
	client := pb.NewErrorMitigatorServiceClient(conn)

	octs := jd.Result.Counts
	zap.L().Debug(fmt.Sprintf("original counts: %v", octs))
	cts := make(map[string]int32)
	shots := int32(0)
	for k, v := range octs {
		cts[k] = int32(v)
		shots += int32(v)
	}
	zap.L().Debug(fmt.Sprintf("pre-mitigation counts: %v", cts))
	dt, err := deviceTopology()
	if err != nil {
		zap.L().Error("failed to get device topology/reason: ", zap.Error(err))
		jd.Status = core.FAILED
		return
	}

	// Convert PhysicalVirtualMapping to sorted MeasuredQubits
	var pvm core.PhysicalVirtualMapping
	if len(jd.Result.TranspilerInfo.PhysicalVirtualMapping) == 0 {
		zap.L().Debug("PhysicalVirtualMapping is nil/use default")
		pvm = make(core.PhysicalVirtualMapping)
		for i := 0; i < len(dt.Qubits); i++ {
			pvm[uint32(i)] = uint32(i)
		}
	} else {
		pvm = jd.Result.TranspilerInfo.PhysicalVirtualMapping
	}
	zap.L().Debug(fmt.Sprintf("PhysicalVirtualMapping: %v", pvm))

	keys := []uint32{}
	for k := range pvm {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	mq := []uint32{}
	for key := range keys {
		mq = append(mq, pvm[uint32(key)])
	}

	mreq := &pb.ReqMitigationRequest{
		DeviceTopology: dt,
		Counts:         cts,
		Shots:          shots, // it is redundant...
		MeasuredQubits: mq,
	}
	zap.L().Debug(fmt.Sprintf("MitigationJob Request: %v", mreq))

	// send request to gRPC server
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	res, err := client.ReqMitigation(ctx, mreq)
	if err != nil {
		zap.L().Error("Failed to request mitigation", zap.Error(err))
		jd.Status = core.FAILED
		return
	}

	zap.L().Debug(fmt.Sprintf("MitigationJob Response: %v", res))
	zap.L().Debug(fmt.Sprintf("MitigationJob Result Counts: %v", res.Counts))
	// get lower bits from counts
	lbcts := make(map[string]uint32)
	for k, v := range res.Counts {
		lbcts[getLowerBits(k, numOfQubits)] = uint32(v)
	}
	zap.L().Debug(fmt.Sprintf("get lower bits of MitigationJob Result Counts: %v", lbcts))
	jd.Result.Counts = lbcts
	jd.Status = core.SUCCEEDED
	zap.L().Debug(fmt.Sprintf("MitigationJob Result: %v", jd.Result))
}

func getNumOfQubits(counts core.Counts) (int, error) {
	if len(counts) == 0 {
		return 0, fmt.Errorf("counts is empty")
	}
	candidateNum := 0
	for k := range counts {
		if candidateNum == 0 {
			candidateNum = len(k)
		} else {
			if candidateNum != len(k) {
				return 0, fmt.Errorf("different length of keys in counts")
			}
		}
	}
	return candidateNum, nil
}

func deviceTopology() (*pb.DeviceTopology, error) {
	s := core.GetSystemComponents()
	disj := s.GetDeviceInfo().DeviceInfoSpecJson
	zap.L().Debug(fmt.Sprintf("device info spec json: %v", disj))

	var dis core.DeviceInfoSpec
	if err := json.Unmarshal([]byte(disj), &dis); err != nil {
		zap.L().Error("failed to unmarshal device info spec json", zap.Error(err))
		return nil, err
	}
	dt := &pb.DeviceTopology{}
	dt.Reset()
	dt.Name = dis.DeviceID
	var qubitsInDeviceTopology []*pb.Qubit
	for _, q := range dis.Qubits {
		mq := pb.Qubit{
			Id:        int32(q.ID),
			GateError: float32(q.Fidelity),
			T1:        float32(q.QubitLife.T1),
			T2:        float32(q.QubitLife.T2),
			MesError: &pb.MesError{
				P0M1: float32(q.MeasError.ProbMeas0Prep1),
				P1M0: float32(q.MeasError.ProbMeas1Prep0),
			},
		}
		qubitsInDeviceTopology = append(qubitsInDeviceTopology, &mq)
	}
	dt.Qubits = qubitsInDeviceTopology
	zap.L().Debug(fmt.Sprintf("device topology: %v", dt))
	return dt, nil
}

func getLowerBits(binStr string, n int) string {
	length := len(binStr)
	if n >= length {
		return binStr
	}
	return binStr[length-n:]
}
