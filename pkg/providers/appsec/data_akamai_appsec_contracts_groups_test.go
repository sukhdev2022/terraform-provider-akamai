package appsec

import (
	"encoding/json"
	"testing"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v2/pkg/appsec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/mock"
)

func TestAccAkamaiContractsGroups_data_basic(t *testing.T) {
	t.Run("match by ContractsGroups ID", func(t *testing.T) {
		client := &mockappsec{}

		getContractsGroupsResponse := appsec.GetContractsGroupsResponse{}
		json.Unmarshal(loadFixtureBytes("testdata/TestDSContractsGroups/ContractsGroups.json"), &getContractsGroupsResponse)

		client.On("GetContractsGroups",
			mock.Anything,
			appsec.GetContractsGroupsRequest{},
		).Return(&getContractsGroupsResponse, nil)

		useClient(client, func() {
			resource.Test(t, resource.TestCase{
				IsUnitTest: true,
				Providers:  testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: loadFixtureString("testdata/TestDSContractsGroups/match_by_id.tf"),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("data.akamai_appsec_contracts_groups.test", "id", "C-1FRYVV3"),
						),
					},
				},
			})
		})

		client.AssertExpectations(t)
	})

}
