package fixtures

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"text/template"

	"crypto/x509/pkix"

	"github.com/paralus/paralus/pkg/sentry/cryptoutil"
	"github.com/paralus/paralus/pkg/service"
	"github.com/rs/xid"
	"sigs.k8s.io/yaml"

	"github.com/paralus/paralus/pkg/log"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	sentry "github.com/paralus/paralus/proto/types/sentry"
	"github.com/shurcooL/httpfs/vfsutil"
)

var _log = log.GetLogger()

var (
	// RelayTemplate is the template for rendering download yaml
	RelayTemplate *template.Template
	// RelayAgentTemplate is the template for rendering download yaml
	RelayAgentTemplate *template.Template
)

// XXX Warning : loadAgentTemplates XXX
// Changes to the below code will affect the default template
// and risks CA cert update. Once the template certs are
// updated following situation will arise.
//   - Existing relay connections will fail, need to restart
//     agents to re-bootstrap.
//   - Existing kubeconfig will fail, need to download new
//     kubeconfig to continue
//
// This can cause issues in the case of production clusters
// Take extra caution while modifying the code to avoid
// unintended side effects.
func loadAgentTemplates(ctx context.Context, bs service.BootstrapService, d map[string]interface{}, pf cryptoutil.PasswordFunc) error {
	yb, err := vfsutil.ReadFile(Fixtures, "agent_templates.yaml")
	if err != nil {
		return err
	}

	t, err := template.New("agent_templates.yaml").Parse(string(yb))
	if err != nil {
		return err
	}

	bb := new(bytes.Buffer)

	err = t.Execute(bb, d)
	if err != nil {
		return err
	}

	var agentTemplateList sentry.BootstrapAgentTemplateList

	jb, err := yaml.YAMLToJSONStrict(bb.Bytes())
	if err != nil {
		return err
	}

	err = json.Unmarshal(jb, &agentTemplateList)
	if err != nil {
		return err
	}

	for _, item := range agentTemplateList.Items {
		// Check bootstrap infr entry exist
		bInfra, _ := bs.GetBootstrapInfra(ctx, item.Spec.InfraRef)
		if bInfra != nil && bInfra.Spec.CaCert != "" {
			// Skip updating bootstrap infra
			_log.Infow("loadAgentTemplates", "skip bootstrap template creation, entry exist", item.Spec.InfraRef)
		} else {
			// Create bootstrap infra entry
			cert, key, err := cryptoutil.GenerateCA(pkix.Name{
				CommonName:         item.Spec.InfraRef,
				Country:            []string{"USA"},
				Organization:       []string{"Paralus"},
				OrganizationalUnit: []string{"Paralus Sentry"},
				Province:           []string{"California"},
				Locality:           []string{"Sunnyvale"},
			}, pf)
			if err != nil {
				return err
			}
			err = bs.PatchBootstrapInfra(ctx, &sentry.BootstrapInfra{
				Metadata: &commonv3.Metadata{
					Name: item.Spec.InfraRef,
				},
				Spec: &sentry.BootstrapInfraSpec{
					CaCert: string(cert),
					CaKey:  string(key),
				},
			})

			if err != nil {
				return err
			}
		}

		item.Spec.Token = xid.New().String()
		// Create/Update bootstrap agent template
		// Token is not updated if entry already exist.
		err = bs.PatchBootstrapAgentTemplate(ctx, item)
		if err != nil {
			return err
		}

	}

	return nil
}

// Load loads fixtures
func Load(ctx context.Context, bs service.BootstrapService, d map[string]interface{}, pf cryptoutil.PasswordFunc) error {
	err := loadAgentTemplates(ctx, bs, d, pf)
	if err != nil {
		return err
	}
	service.KEKFunc = pf

	err = loadRelayTemplate()

	return err
}

func loadRelayTemplate() error {
	f, err := Fixtures.Open("relay_template.yaml")
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	RelayTemplate, err = template.New("relay_template").Parse(string(b))
	if err != nil {
		return err
	}

	f1, err := Fixtures.Open("relay_agent_template.yaml")
	if err != nil {
		return err
	}

	b1, err := ioutil.ReadAll(f1)
	if err != nil {
		return err
	}

	RelayAgentTemplate, err = template.New("relay_agent_template").Parse(string(b1))
	return err
}
