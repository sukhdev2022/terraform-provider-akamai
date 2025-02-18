package appsec

import (
	"encoding/json"
	"testing"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v2/pkg/appsec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/mock"
)

func TestAccAkamaiMatchTargetSequence_res_basic(t *testing.T) {
	t.Run("match by MatchTargetSequence ID", func(t *testing.T) {
		client := &mockappsec{}

		updateSequenceResponse := appsec.UpdateMatchTargetSequenceResponse{}
		json.Unmarshal(loadFixtureBytes("testdata/TestResMatchTargetSequence/MatchTargetSequenceResp.json"), &updateSequenceResponse)

		getSequenceResponse := appsec.GetMatchTargetSequenceResponse{}
		json.Unmarshal(loadFixtureBytes("testdata/TestResMatchTargetSequence/MatchTargetSequence.json"), &getSequenceResponse)

		config := appsec.GetConfigurationResponse{}
		json.Unmarshal(loadFixtureBytes("testdata/TestResConfiguration/LatestConfiguration.json"), &config)

		client.On("GetConfiguration",
			mock.Anything,
			appsec.GetConfigurationRequest{ConfigID: 43253},
		).Return(&config, nil)

		client.On("GetMatchTargetSequence",
			mock.Anything,
			appsec.GetMatchTargetSequenceRequest{ConfigID: 43253, ConfigVersion: 7, Type: "website"},
		).Return(&getSequenceResponse, nil)

		client.On("UpdateMatchTargetSequence",
			mock.Anything,
			appsec.UpdateMatchTargetSequenceRequest{ConfigID: 43253, ConfigVersion: 7, Type: "website", TargetSequence: []appsec.MatchTargetItem{{Sequence: 1, TargetID: 2052813}, {Sequence: 2, TargetID: 2971336}}},
		).Return(&updateSequenceResponse, nil)

		useClient(client, func() {
			resource.Test(t, resource.TestCase{
				IsUnitTest: true,
				Providers:  testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: loadFixtureString("testdata/TestResMatchTargetSequence/match_by_id.tf"),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("akamai_appsec_match_target_sequence.test", "id", "43253:website"),
						),
					},
					{
						Config: loadFixtureString("testdata/TestResMatchTargetSequence/update_by_id.tf"),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("akamai_appsec_match_target_sequence.test", "id", "43253:website"),
						),
					},
				},
			})
		})

		client.AssertExpectations(t)
	})

}
