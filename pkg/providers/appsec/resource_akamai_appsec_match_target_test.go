package appsec

import (
	"encoding/json"
	"testing"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v2/pkg/appsec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/mock"
)

func TestAccAkamaiMatchTarget_res_basic(t *testing.T) {
	t.Run("match by MatchTarget ID", func(t *testing.T) {
		client := &mockappsec{}

		updateMatchTargetResponse := appsec.UpdateMatchTargetResponse{}
		json.Unmarshal(loadFixtureBytes("testdata/TestResMatchTarget/MatchTargetUpdated.json"), &updateMatchTargetResponse)

		getMatchTargetResponse := appsec.GetMatchTargetResponse{}
		json.Unmarshal(loadFixtureBytes("testdata/TestResMatchTarget/MatchTarget.json"), &getMatchTargetResponse)

		getMatchTargetResponseAfterUpdate := appsec.GetMatchTargetResponse{}
		json.Unmarshal(loadFixtureBytes("testdata/TestResMatchTarget/MatchTargetUpdated.json"), &getMatchTargetResponseAfterUpdate)

		createMatchTargetResponse := appsec.CreateMatchTargetResponse{}
		json.Unmarshal(loadFixtureBytes("testdata/TestResMatchTarget/MatchTargetCreated.json"), &createMatchTargetResponse)

		removeMatchTargetResponse := appsec.RemoveMatchTargetResponse{}
		json.Unmarshal(loadFixtureBytes("testdata/TestResMatchTarget/MatchTargetCreated.json"), &removeMatchTargetResponse)

		config := appsec.GetConfigurationResponse{}
		json.Unmarshal(loadFixtureBytes("testdata/TestResConfiguration/LatestConfiguration.json"), &config)

		client.On("GetConfiguration",
			mock.Anything,
			appsec.GetConfigurationRequest{ConfigID: 43253},
		).Return(&config, nil)

		client.On("GetMatchTarget",
			mock.Anything,
			appsec.GetMatchTargetRequest{ConfigID: 43253, ConfigVersion: 7, TargetID: 3008967},
		).Return(&getMatchTargetResponse, nil).Times(3)

		client.On("GetMatchTarget",
			mock.Anything,
			appsec.GetMatchTargetRequest{ConfigID: 43253, ConfigVersion: 7, TargetID: 3008967},
		).Return(&getMatchTargetResponseAfterUpdate, nil)

		createMatchTargetJSON := loadFixtureBytes("testdata/TestResMatchTarget/CreateMatchTarget.json")
		client.On("CreateMatchTarget",
			mock.Anything,
			appsec.CreateMatchTargetRequest{Type: "", ConfigID: 43253, ConfigVersion: 7, JsonPayloadRaw: createMatchTargetJSON},
		).Return(&createMatchTargetResponse, nil)

		updateMatchTargetJSON := loadFixtureBytes("testdata/TestResMatchTarget/UpdateMatchTarget.json")
		client.On("UpdateMatchTarget",
			mock.Anything,
			appsec.UpdateMatchTargetRequest{ConfigID: 43253, ConfigVersion: 7, TargetID: 3008967, JsonPayloadRaw: updateMatchTargetJSON},
		).Return(&updateMatchTargetResponse, nil)

		client.On("RemoveMatchTarget",
			mock.Anything,
			appsec.RemoveMatchTargetRequest{ConfigID: 43253, ConfigVersion: 7, TargetID: 3008967},
		).Return(&removeMatchTargetResponse, nil)

		useClient(client, func() {
			resource.Test(t, resource.TestCase{
				IsUnitTest: true,
				Providers:  testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: loadFixtureString("testdata/TestResMatchTarget/match_by_id.tf"),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("akamai_appsec_match_target.test", "id", "43253:3008967"),
						),
					},
					{
						Config: loadFixtureString("testdata/TestResMatchTarget/update_by_id.tf"),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("akamai_appsec_match_target.test", "id", "43253:3008967"),
						),
					},
				},
			})
		})

		client.AssertExpectations(t)
	})

}
