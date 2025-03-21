package core

import (
	"context"
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/oklog/run"
	"github.com/oqtopus-team/oqtopus-engine/coreapp/common"
	"go.uber.org/zap"
)

var runContext *RunContext

const (
	PERIODIC_TASKS       = "periodic_tasks"
	INTERNAL_JOB_SERVERS = "internal_job_servers"
	API_SERVERS          = "api_servers"
)

type PeriodicTaskImplMap map[string]PeriodicTaskImpl
type InternalJobServerImplMap map[string]InternalJobServerImpl
type APIServerImplMap map[string]APIServerImpl

type PeriodicTaskMap map[string]*PeriodicTask
type InternalJobServerMap map[string]*InternalJobServer
type APIServerMap map[string]*APIServer

type ImplMaps struct {
	PeriodicTaskImplMap      PeriodicTaskImplMap
	InternalJobServerImplMap InternalJobServerImplMap
	APIServerImplMap         APIServerImplMap
}

type Runner interface {
	*PeriodicTask | *InternalJobServer | *APIServer
	GetParams() interface{}
}

type RunnerImpl interface {
	GetEmptyParams() interface{}
	SetParams(interface{}) error
	Setup() error
}

type RunContext struct {
	*run.Group
	context.Context

	settingsPath string

	RunGroupMaps *RunGroupMaps `toml:"run_group,omitempty"`
}

// This will be replaced with "setting"
type RungroupSetting struct {
	Entries map[string]interface{} `toml:"run_group,omitempty"`
}

func NewGroupSettings() *RungroupSetting {
	return &RungroupSetting{
		Entries: make(map[string]interface{}),
	}
}

type RunGroupMaps struct {
	PeriodicTasks      PeriodicTaskMap      `toml:"periodic_tasks"`
	InternalJobServers InternalJobServerMap `toml:"internal_job_servers"`
	APIServers         APIServerMap         `toml:"api_servers"`
}

func parseRunGroupSettings(settings map[string]interface{}, im *ImplMaps) (*RunGroupMaps, error) {
	rgm := &RunGroupMaps{
		PeriodicTasks:      make(PeriodicTaskMap),
		InternalJobServers: make(InternalJobServerMap),
		APIServers:         make(APIServerMap),
	}
	for group, value := range settings {
		switch group {
		case PERIODIC_TASKS:
			zap.L().Debug(fmt.Sprintf("PeriodicTasks: %v", value))
			ptm, err := parseRunnerSettings[*PeriodicTask, PeriodicTaskImpl](value.(map[string]interface{}), im.PeriodicTaskImplMap)
			if err != nil {
				zap.L().Error(fmt.Sprintf("Failed to parse periodic tasks settings. Reason:%s", err))
				return nil, err
			}
			rgm.PeriodicTasks = ptm
		case INTERNAL_JOB_SERVERS:
			zap.L().Debug(fmt.Sprintf("InternalJobServers: %v", value))
			ijs, err := parseRunnerSettings[*InternalJobServer, InternalJobServerImpl](value.(map[string]interface{}), im.InternalJobServerImplMap)
			if err != nil {
				zap.L().Error(fmt.Sprintf("Failed to parse internal job servers settings. Reason:%s", err))
				return nil, err
			}
			rgm.InternalJobServers = ijs
		case API_SERVERS:
			zap.L().Debug(fmt.Sprintf("APIServers: %v", value))
			asm, err := parseRunnerSettings[*APIServer, APIServerImpl](value.(map[string]interface{}), im.APIServerImplMap)
			if err != nil {
				zap.L().Error(fmt.Sprintf("Failed to parse api servers settings. Reason:%s", err))
				return nil, err
			}
			rgm.APIServers = asm
		default:
			msg := fmt.Sprintf("Unknown run group type. Group:%s, Value:%v", group, value)
			zap.L().Error(msg)
			return nil, fmt.Errorf(msg)
		}
	}
	zap.L().Debug("Successfully parsed run group settings.", zap.Any("RunGroupMaps", rgm))
	return rgm, nil
}

func parseRunnerSettings[R Runner, I RunnerImpl](settings map[string]interface{}, implMap map[string]I) (map[string]R, error) {
	runnerMap := make(map[string]R)
	for k, v := range implMap {
		zap.L().Debug(fmt.Sprintf("implMap/key:%s/value:%v", k, v))
	}
	for runnerName := range settings { // value is not used for now
		impl, ok := implMap[runnerName]
		if !ok {
			msg := fmt.Sprintf("failed to find %s implementation from RunnerMap %v", runnerName, implMap)
			zap.L().Error(msg)
			return nil, fmt.Errorf(msg)
		}
		runner, err := newRunner[R, I](impl)
		if err != nil {
			msg := fmt.Sprintf("failed to set implementation to runnerName:%v/impl:%v/reason:%v", runnerName, impl, err.Error())
			zap.L().Error(msg)
			return nil, fmt.Errorf(msg)
		}
		runnerMap[runnerName] = runner
	}
	return runnerMap, nil
}

func newRunner[R Runner, I RunnerImpl](runnerImpl I) (runner R, err error) {
	err = nil
	switch any(runner).(type) {
	case *PeriodicTask:
		i, ok := any(runnerImpl).(PeriodicTaskImpl)
		if !ok {
			err = fmt.Errorf("failed to cast to PeriodicTaskImpl/runner:%v", runner)
			return
		}
		runner = any(&PeriodicTask{PeriodicTaskImpl: i}).(R)
	case *InternalJobServer:
		i, ok := any(runnerImpl).(InternalJobServerImpl)
		if !ok {
			err = fmt.Errorf("failed to cast to InternalJobServerImpl/runner:%v", runner)
			return
		}
		runner = any(&InternalJobServer{InternalJobServerImpl: i}).(R)
	case *APIServer:
		i, ok := any(runnerImpl).(APIServerImpl)
		if !ok {
			err = fmt.Errorf("failed to cast to APIServerImpl/runner:%v", runner)
			return
		}
		runner = any(&APIServer{APIServerImpl: i}).(R)
	default:
		err = fmt.Errorf("unknown runner type:%v", runner)
		return
	}
	return
}

func NewRunContext() *RunContext {
	return &RunContext{
		Group:   &run.Group{},
		Context: context.Background(),
		RunGroupMaps: &RunGroupMaps{
			PeriodicTasks:      make(PeriodicTaskMap),
			InternalJobServers: make(InternalJobServerMap),
			APIServers:         make(APIServerMap),
		},
	}
}

// TODO: refactor this too long function
// TODO: a lot of tests are needed
func NewRunContextWithSettingPath(settingsPath string, im *ImplMaps) (*RunContext, error) {
	tomlString, err := common.ReadSettingsFile(settingsPath)
	if err != nil {
		zap.L().Error(fmt.Sprintf("failed to read settings file/reason:%s", err))
		return nil, err
	}
	// Decoding TOML to RunGroupMaps is a bit tricky.
	// 1. decode to Settings to get RunGroupSettings
	// It just decodes to setup RunGroupMaps
	s := NewGroupSettings()
	if metadata, err := toml.Decode(tomlString, s); err != nil {
		zap.L().Error(fmt.Sprintf("Failed to decode settings file. Reason:%s. Metadata:%v",
			err, metadata))
		return nil, err
	}
	zap.L().Debug("Successfully decoded TOML file to Settings.", zap.Any("Settings", s))
	runGroupMaps, err := parseRunGroupSettings(s.Entries, im)
	if err != nil {
		zap.L().Error(fmt.Sprintf("Failed to parse run group settings. Reason:%s", err))
		return nil, err
	}
	zap.L().Debug("Successfully parsed run group settings.", zap.Any("RunGroupMaps", runGroupMaps))
	// 2. decode to RunGroupMaps
	rc := &RunContext{
		Group:        &run.Group{},
		Context:      context.Background(),
		settingsPath: settingsPath,
		RunGroupMaps: runGroupMaps,
	}
	// 3. store Impl to tmp map,
	// because we need to recover them after decoding to RunGroupMaps
	tmpPeriodicTaskImplMap := make(map[string]PeriodicTaskImpl)
	tmpInternalJobServerImplMap := make(map[string]InternalJobServerImpl)
	tmpAPIServerImplMap := make(map[string]APIServerImpl)
	for taskName, task := range rc.RunGroupMaps.PeriodicTasks {
		tmpPeriodicTaskImplMap[taskName] = task.PeriodicTaskImpl
	}
	for serverName, server := range rc.RunGroupMaps.InternalJobServers {
		tmpInternalJobServerImplMap[serverName] = server.InternalJobServerImpl
	}
	for serverName, server := range rc.RunGroupMaps.APIServers {
		tmpAPIServerImplMap[serverName] = server.APIServerImpl
	}
	// 4. decode to RunGroupMaps
	if metadata, err := toml.Decode(string(tomlString), rc); err != nil {
		zap.L().Error(fmt.Sprintf("Failed to decode settings file. Reason:%s. Metadata:%v",
			err, metadata))
		return nil, err
	}
	zap.L().Debug("Successfully decoded TOML file to RunGroupMaps.", zap.Any("RunGroupMaps", rc.RunGroupMaps))
	// 5. recover Impl
	for taskName, task := range rc.RunGroupMaps.PeriodicTasks {
		task.PeriodicTaskImpl = tmpPeriodicTaskImplMap[taskName]
	}
	for serverName, server := range rc.RunGroupMaps.InternalJobServers {
		server.InternalJobServerImpl = tmpInternalJobServerImplMap[serverName]
	}
	for serverName, server := range rc.RunGroupMaps.APIServers {
		server.APIServerImpl = tmpAPIServerImplMap[serverName]
	}
	zap.L().Debug("Successfully recovered PeriodicTasks Impl and InternalJobServer Impl.",
		zap.Any("RunGroupMaps", rc.RunGroupMaps))
	// 6. set parmeters to Impl
	if err := setParametersToImpl[*PeriodicTask](rc.RunGroupMaps.PeriodicTasks); err != nil {
		zap.L().Error(fmt.Sprintf("failed to set parameters to PeriodicTask Impl/reason:%s", err.Error()))
		return nil, err
	}
	if err := setParametersToImpl[*InternalJobServer](rc.RunGroupMaps.InternalJobServers); err != nil {
		zap.L().Error(fmt.Sprintf("failed to set parameters to InternalJobServer Impl/reason:%s", err.Error()))
		return nil, err
	}
	if err := setParametersToImpl[*APIServer](rc.RunGroupMaps.APIServers); err != nil {
		zap.L().Error(fmt.Sprintf("failed to set parameters to APIServers Impl/reason:%s", err.Error()))
		return nil, err
	}

	zap.L().Debug("Successfully set parameters to Impl.",
		zap.Any("RunGroupMaps", rc.RunGroupMaps))
	// 7. setup Impl and add to RunContext
	if err := setupImplAndAddToRunContext[*PeriodicTask](rc.RunGroupMaps.PeriodicTasks, rc.AddPeriodicTask); err != nil {
		zap.L().Error(fmt.Sprintf("failed to setup and add PeriodicTask/reason:%s", err.Error()))
		return nil, err
	}
	if err := setupImplAndAddToRunContext[*InternalJobServer](rc.RunGroupMaps.InternalJobServers, rc.AddInternalJobServer); err != nil {
		zap.L().Error(fmt.Sprintf("failed to setup and add InternalJobServer/reason:%s", err.Error()))
		return nil, err
	}
	if err := setupImplAndAddToRunContext[*APIServer](rc.RunGroupMaps.APIServers, rc.AddAPIServer); err != nil {
		zap.L().Error(fmt.Sprintf("failed to setup and add APIServer/reason:%s", err.Error()))
		return nil, err
	}

	zap.L().Info("Successfully initialized RunContext. RunGroupMaps:", zap.Any("RunGroupMaps", rc.RunGroupMaps))
	return rc, nil
}

func setParametersToImpl[R Runner](runners map[string]R) error {
	for name, runner := range runners {
		zap.L().Debug(fmt.Sprintf("setting parameters to Impl/name:%s/runner%v", name, runner))
		if err := any(runner).(RunnerImpl).SetParams(runner.GetParams()); err != nil {
			zap.L().Error(fmt.Sprintf("failed to set parameters to Impl/name:%s/runner%v/reason:%s",
				name, runner, err.Error()))
			return err
		}
	}
	return nil
}

func setupImplAndAddToRunContext[R Runner](
	runners map[string]R,
	addFunc func(R, string) error) error {
	for name, runner := range runners {
		if err := any(runner).(RunnerImpl).Setup(); err != nil {
			zap.L().Error(fmt.Sprintf("failed to setup/name:%s/reason:%s", name, err.Error()))
			return err
		}
		if err := addFunc(runner, name); err != nil {
			zap.L().Error(fmt.Sprintf("failed to add runner/name:%s/reason:%s", name, err))
			return err
		}
		zap.L().Info(fmt.Sprintf("successfully added runner/name:%s", name))
	}
	return nil
}

func GetRunContext() *RunContext {
	return runContext
}

func SetRunContext(rc *RunContext) {
	runContext = rc
}

type PeriodicTask struct {
	Period time.Duration `toml:"period"`
	Params interface{}   `toml:"params,omitempty"`
	PeriodicTaskImpl
}

func (t *PeriodicTask) GetParams() interface{} {
	return t.Params
}

type PeriodicTaskImpl interface {
	RunnerImpl
	RequirePeriodUpdate() (ok bool, duration time.Duration)
	Task()
	Cleanup()
}

type DefaultTaskImpl struct{}

func (v *DefaultTaskImpl) Setup() error {
	return nil
}

func (v *DefaultTaskImpl) GetEmptyParams() interface{} {
	return v
}

func (v *DefaultTaskImpl) SetParams(p interface{}) error {
	return nil
}

func (v *DefaultTaskImpl) RequirePeriodUpdate() (bool, time.Duration) {
	return false, 0
}

func (v *DefaultTaskImpl) Task() {}

func (v *DefaultTaskImpl) Cleanup() {}

func (rc *RunContext) AddPeriodicTask(t *PeriodicTask, taskName string) error {
	ctx, cancel := context.WithCancel(rc.Context)
	lastPeriod := t.Period
	rc.Group.Add(
		func() error {
			ticker := time.NewTicker(t.Period)
			zap.L().Info(fmt.Sprintf("[PeriodicTask/%s/Start]", taskName))
			t.PeriodicTaskImpl.Task()
			for {
				select {
				case <-ctx.Done():
					zap.L().Info(fmt.Sprintf("[PeriodicTask/%s/TearDown]Cleaning up periodic task", taskName))
					ticker.Stop()
					t.PeriodicTaskImpl.Cleanup()
					zap.L().Info(fmt.Sprintf("[PeriodicTask/%s/TearDown]Cleaned up periodic task", taskName))
					return ctx.Err()
				case <-ticker.C:
					t.PeriodicTaskImpl.Task()
					ok, newPeriod := t.RequirePeriodUpdate()
					if ok && newPeriod != lastPeriod {
						zap.L().Info(fmt.Sprintf("[PeriodicTask/%s/ResetPeriod]Resetting periodic task. from %v to %v",
							taskName, lastPeriod, newPeriod))
						ticker.Reset(newPeriod)
						zap.L().Info(fmt.Sprintf("[PeriodicTask/%s/ResetPeriod]Reset periodic task. from %v to %v",
							taskName, lastPeriod, newPeriod))
						lastPeriod = newPeriod
					}
				}
			}
		},
		func(error) {
			zap.L().Info(fmt.Sprintf("[PeriodicTask/%s/TearDown]Cancelling periodic task", taskName))
			cancel()
			zap.L().Info(fmt.Sprintf("[PeriodicTask/%s/TearDown]Canceled periodic task", taskName))
		},
	)
	return nil
}

type InternalJobServer struct {
	Params interface{} `toml:"params,omitempty"`
	InternalJobServerImpl
}

func (s *InternalJobServer) GetParams() interface{} {
	return s.Params
}

func (s *InternalJobServer) SetImpl(impl interface{}) error {
	ijs, ok := impl.(InternalJobServerImpl)
	if !ok {
		msg := fmt.Sprintf("Failed to cast to InternalJobServerImpl. Impl:%v", impl)
		zap.L().Error(msg)
		return fmt.Errorf(msg)
	}
	s.InternalJobServerImpl = ijs
	return nil
}

type InternalJobServerImpl interface {
	RunnerImpl
	Start() error
	Cleanup()
	Handle(Job) error
}

func NewInternalJobServer(impl InternalJobServerImpl) *InternalJobServer {
	return &InternalJobServer{
		Params:                impl.GetEmptyParams(),
		InternalJobServerImpl: impl,
	}
}

func (rc *RunContext) AddInternalJobServer(s *InternalJobServer, serverName string) error {
	ctx, cancel := context.WithCancel(rc.Context)
	rc.Group.Add(
		func() error {
			zap.L().Info(fmt.Sprintf("[InternalJobServer/%s/Start]", serverName))
			err := s.Start()
			if err != nil {
				zap.L().Error(fmt.Sprintf("[InternalJobServer/%s/Error]failed to start internal job server/reason:%s",
					serverName, err))
				return err
			}
			zap.L().Info(fmt.Sprintf("[InternalJobServer/%s/Started]", serverName))
			<-ctx.Done()
			zap.L().Info(fmt.Sprintf("[InternalJobServer/%s/TearDown]cleaning up internal job server",
				serverName))
			s.Cleanup()
			zap.L().Info(fmt.Sprintf("[InternalJobServer/%s/TearDown]cleaned up internal job server",
				serverName))
			return nil
		},
		func(error) {
			zap.L().Info(fmt.Sprintf("[InternalJobServer/%s/TearDown]cancelling internal job server",
				serverName))
			cancel()
			zap.L().Info(fmt.Sprintf("[InternalJobServer/%s/TearDown]canceled internal job server",
				serverName))
		},
	)
	return nil
}

type APIServer struct {
	Params interface{} `toml:"params,omitempty"`
	APIServerImpl
}

func (s *APIServer) GetParams() interface{} {
	return s.Params
}

type APIServerImpl interface {
	RunnerImpl
	Serve() error
	Shutdown()
}

func NewAPIServer(impl APIServerImpl) *APIServer {
	return &APIServer{
		Params:        impl.GetEmptyParams(),
		APIServerImpl: impl,
	}
}

func (rc *RunContext) AddAPIServer(s *APIServer, serverName string) error {
	rc.Group.Add(
		func() error {
			zap.L().Info(fmt.Sprintf("[APIServer/%s/Start]", serverName))
			if err := s.Serve(); err != nil {
				zap.L().Error(fmt.Sprintf("[APIServer/%s/Error]failed to start api server/reason:%s",
					serverName, err.Error()))
				return err
			}
			return nil
		},
		func(error) {
			zap.L().Info(fmt.Sprintf("[APIServer/%s/TearDown]shutting down api server", serverName))
			s.Shutdown()
			zap.L().Info(fmt.Sprintf("[APIServer/%s/TearDown]shut down api server", serverName))
		},
	)
	return nil
}
