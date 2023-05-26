// Package nsm - Common operations for the adapter
package nsm

import (
	"context"
	"fmt"

	"github.com/layer5io/meshery-adapter-library/adapter"
	"github.com/layer5io/meshery-adapter-library/common"
	adapterconfig "github.com/layer5io/meshery-adapter-library/config"
	"github.com/layer5io/meshery-adapter-library/status"
	internalconfig "github.com/layer5io/meshery-nsm/internal/config"
	"github.com/layer5io/meshkit/errors"
	"github.com/layer5io/meshkit/logger"
)

const (
	// SMIManifest is the manifest.yaml file for smi conformance tool
	SMIManifest = "https://raw.githubusercontent.com/layer5io/learn-layer5/master/smi-conformance/manifest.yml"
)

// Mesh represents the nsm-mesh adapter and embeds adapter.Adapter
type Mesh struct {
	adapter.Adapter // Type Embedded
}

// New initializes treafik-mesh handler.
func New(c adapterconfig.Handler, l logger.Handler, kc adapterconfig.Handler) adapter.Handler {
	return &Mesh{
		Adapter: adapter.Adapter{
			Config:            c,
			Log:               l,
			KubeconfigHandler: kc,
		},
	}
}

// ApplyOperation applies the operation on nsm mesh
func (mesh *Mesh) ApplyOperation(ctx context.Context, opReq adapter.OperationRequest) error {
	operations := make(adapter.Operations)
	err := mesh.Config.GetObject(adapter.OperationsKey, &operations)
	if err != nil {
		return err
	}

	e := &adapter.Event{
		OperationId: opReq.OperationID,
		Summary:     status.Deploying,
		Details:     "Operation is not supported",
		Component:   internalconfig.ServerConfig["type"],
		ComponentName: internalconfig.ServerConfig["name"],
	}

	switch opReq.OperationName {
	case internalconfig.NSMMeshOperation:
		go func(hh *Mesh, ee *adapter.Event) {
			version := string(operations[opReq.OperationName].Versions[0])
			stat, err := hh.installNSMMesh(opReq.IsDeleteOperation, version, opReq.Namespace)
			if err != nil {
				summary := fmt.Sprintf("Error while %s NSM service mesh", stat)
				hh.streamErr(summary, e, err)
				return
			}
			ee.Summary = fmt.Sprintf("NSM service mesh %s successfully", stat)
			ee.Details = fmt.Sprintf("The NSM service mesh is now %s.", stat)
			hh.StreamInfo(e)
		}(mesh, e)
	case common.BookInfoOperation, common.HTTPBinOperation, common.ImageHubOperation, common.EmojiVotoOperation:
		go func(hh *Mesh, ee *adapter.Event) {
			appName := operations[opReq.OperationName].AdditionalProperties[common.ServiceName]
			stat, err := hh.installSampleApp(opReq.Namespace, opReq.IsDeleteOperation, operations[opReq.OperationName].Templates)
			if err != nil {
				summary := fmt.Sprintf("Error while %s %s application", stat, appName)
				hh.streamErr(summary, e, err)
				return
			}
			ee.Summary = fmt.Sprintf("%s application %s successfully", appName, stat)
			ee.Details = fmt.Sprintf("The %s application is now %s.", appName, stat)
			hh.StreamInfo(e)
		}(mesh, e)
	case common.CustomOperation:
		go func(hh *Mesh, ee *adapter.Event) {
			stat, err := hh.applyCustomOperation(opReq.Namespace, opReq.CustomBody, opReq.IsDeleteOperation)
			if err != nil {
				summary := fmt.Sprintf("Error while %s custom operation", stat)
				hh.streamErr(summary, e, err)
				return
			}
			ee.Summary = fmt.Sprintf("Manifest %s successfully", status.Deployed)
			ee.Details = ""
			hh.StreamInfo(e)
		}(mesh, e)
	case internalconfig.NSMICMPResponderSampleApp, internalconfig.NSMVPPICMPResponderSampleApp, internalconfig.NSMVPMSampleApp:
		go func(hh *Mesh, ee *adapter.Event) {
			version := string(operations[internalconfig.NSMMeshOperation].Versions[0])
			chart := operations[opReq.OperationName].AdditionalProperties[internalconfig.HelmChart]
			appName := operations[opReq.OperationName].AdditionalProperties[common.ServiceName]

			stat, err := hh.installNSMSampleApp(opReq.Namespace, chart, version, opReq.IsDeleteOperation)
			if err != nil {
				summary := fmt.Sprintf("Error while %s %s application", stat, appName)
				hh.streamErr(summary, e, err)
				return
			}
			ee.Summary = fmt.Sprintf("%s application %s successfully", appName, stat)
			ee.Details = fmt.Sprintf("The %s application is now %s.", appName, stat)
			hh.StreamInfo(e)
		}(mesh, e)
	case common.SmiConformanceOperation:
		go func(hh *Mesh, ee *meshes.EventsResponse) {
			name := operations[opReq.OperationName].Description
			_, err := hh.RunSMITest(adapter.SMITestOptions{
				Ctx:         context.TODO(),
				OperationID: ee.OperationId,
				Manifest:    SMIManifest,
				Namespace:   "meshery",
				Labels:      make(map[string]string),
				Annotations: make(map[string]string),
			})
			if err != nil {
				summary := fmt.Sprintf("Error while %s %s test", status.Running, name)
				hh.streamErr(summary, e, err)
				return
			}
			ee.Summary = fmt.Sprintf("%s test %s successfully", name, status.Completed)
			ee.Details = ""
			hh.StreamInfo(e)
		}(mesh, e)
	default:
		mesh.streamErr("Invalid operation", e, ErrOpInvalid)
	}

	return nil
}

func(mesh *Mesh) streamErr(summary string, e *adapter.Event, err error) {
	e.Summary = summary
	e.Details = err.Error()
	e.ErrorCode = errors.GetCode(err)
	e.ProbableCause = errors.GetCause(err)
	e.SuggestedRemediation = errors.GetRemedy(err)
	mesh.StreamErr(e, err)
}
