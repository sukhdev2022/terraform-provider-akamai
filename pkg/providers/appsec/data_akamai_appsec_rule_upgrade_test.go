package appsec

import (
	"encoding/json"
	"testing"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v2/pkg/appsec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/mock"
)

func TestAccAkamaiRuleUpgrade_data_basic(t *testing.T) {
	t.Run("match by RuleUpgrade ID", func(t *testing.T) {
		client := &mockappsec{}

		getRuleUpgradeResponse := appsec.GetRuleUpgradeResponse{}
		json.Unmarshal(loadFixtureBytes("testdata/TestDSRuleUpgrade/RuleUpgrade.json"), &getRuleUpgradeResponse)

		config := appsec.GetConfigurationResponse{}
		json.Unmarshal(loadFixtureBytes("testdata/TestResConfiguration/LatestConfiguration.json"), &config)

		client.On("GetConfiguration",
			mock.Anything,
			appsec.GetConfigurationRequest{ConfigID: 43253},
		).Return(&config, nil)

		client.On("GetRuleUpgrade",
			mock.Anything,
			appsec.GetRuleUpgradeRequest{ConfigID: 43253, Version: 7, PolicyID: "AAAA_81230"},
		).Return(&getRuleUpgradeResponse, nil)

		useClient(client, func() {
			resource.Test(t, resource.TestCase{
				IsUnitTest: true,
				Providers:  testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: loadFixtureString("testdata/TestDSRuleUpgrade/match_by_id.tf"),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("data.akamai_appsec_rule_upgrade_details.test", "id", "43253"),
						),
					},
				},
			})
		})

		client.AssertExpectations(t)
	})

}
