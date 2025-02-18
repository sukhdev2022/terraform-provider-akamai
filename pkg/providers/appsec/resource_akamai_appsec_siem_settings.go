package appsec

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v2/pkg/appsec"
	"github.com/akamai/terraform-provider-akamai/v2/pkg/akamai"
	"github.com/akamai/terraform-provider-akamai/v2/pkg/tools"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// appsec v1
//
// https://techdocs.akamai.com/application-security/reference/api
func resourceSiemSettings() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSiemSettingsCreate,
		ReadContext:   resourceSiemSettingsRead,
		UpdateContext: resourceSiemSettingsUpdate,
		DeleteContext: resourceSiemSettingsDelete,
		CustomizeDiff: customdiff.All(
			VerifyIDUnchanged,
		),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"config_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"enable_siem": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"enable_for_all_policies": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"security_policy_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"enable_botman_siem": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"siem_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func resourceSiemSettingsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := akamai.Meta(m)
	client := inst.Client(meta)
	logger := meta.Log("APPSEC", "resourceSiemSettingsCreate")
	logger.Debugf("in resourceSiemSettingsCreate")

	configID, err := tools.GetIntValue("config_id", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}
	version, err := getModifiableConfigVersion(ctx, configID, "siemSetting", m)
	if err != nil {
		return diag.FromErr(err)
	}
	enableSiem, err := tools.GetBoolValue("enable_siem", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}
	enableForAllPolicies, err := tools.GetBoolValue("enable_for_all_policies", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}
	securityPolicyIDs, err := tools.GetSetValue("security_policy_ids", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}
	spIDs := make([]string, 0, len(securityPolicyIDs.List()))
	for _, h := range securityPolicyIDs.List() {
		spIDs = append(spIDs, h.(string))

	}
	enableBotmanSiem, err := tools.GetBoolValue("enable_botman_siem", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}
	siemID, err := tools.GetIntValue("siem_id", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}

	createSiemSettings := appsec.UpdateSiemSettingsRequest{
		ConfigID:                configID,
		Version:                 version,
		EnableSiem:              enableSiem,
		EnableForAllPolicies:    enableForAllPolicies,
		FirewallPolicyIds:       spIDs,
		EnabledBotmanSiemEvents: enableBotmanSiem,
		SiemDefinitionID:        siemID,
	}

	_, err = client.UpdateSiemSettings(ctx, createSiemSettings)
	if err != nil {
		logger.Errorf("calling 'createSiemSettings': %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d", createSiemSettings.ConfigID))

	return resourceSiemSettingsRead(ctx, d, m)
}

func resourceSiemSettingsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := akamai.Meta(m)
	client := inst.Client(meta)
	logger := meta.Log("APPSEC", "resourceSiemSettingsRead")
	logger.Debugf("in resourceSiemSettingsRead")

	configID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	version, err := getLatestConfigVersion(ctx, configID, m)
	if err != nil {
		return diag.FromErr(err)
	}

	getSiemSettings := appsec.GetSiemSettingsRequest{
		ConfigID: configID,
		Version:  version,
	}

	siemsettings, err := client.GetSiemSettings(ctx, getSiemSettings)
	if err != nil {
		logger.Errorf("calling 'getSiemSettings': %s", err.Error())
		return diag.FromErr(err)
	}

	if err := d.Set("config_id", getSiemSettings.ConfigID); err != nil {
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}
	if err := d.Set("enable_siem", siemsettings.EnableSiem); err != nil {
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}
	if err := d.Set("enable_for_all_policies", siemsettings.EnableForAllPolicies); err != nil {
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}
	if err := d.Set("security_policy_ids", siemsettings.FirewallPolicyIds); err != nil {
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}
	if err := d.Set("enable_botman_siem", siemsettings.EnabledBotmanSiemEvents); err != nil {
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}
	if err := d.Set("siem_id", siemsettings.SiemDefinitionID); err != nil {
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}

	return nil
}

func resourceSiemSettingsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := akamai.Meta(m)
	client := inst.Client(meta)
	logger := meta.Log("APPSEC", "resourceSiemSettingsUpdate")
	logger.Debugf("in resourceSiemSettingsUpdate")

	configID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	version, err := getModifiableConfigVersion(ctx, configID, "siemSetting", m)
	if err != nil {
		return diag.FromErr(err)
	}
	enableSiem, err := tools.GetBoolValue("enable_siem", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}
	enableForAllPolicies, err := tools.GetBoolValue("enable_for_all_policies", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}
	securityPolicyIDs, err := tools.GetSetValue("security_policy_ids", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}
	spIDs := make([]string, 0, len(securityPolicyIDs.List()))
	for _, h := range securityPolicyIDs.List() {
		spIDs = append(spIDs, h.(string))

	}
	enableBotmanSiem, err := tools.GetBoolValue("enable_botman_siem", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}
	siemID, err := tools.GetIntValue("siem_id", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}

	updateSiemSettings := appsec.UpdateSiemSettingsRequest{
		ConfigID:                configID,
		Version:                 version,
		EnableSiem:              enableSiem,
		EnableForAllPolicies:    enableForAllPolicies,
		FirewallPolicyIds:       spIDs,
		EnabledBotmanSiemEvents: enableBotmanSiem,
		SiemDefinitionID:        siemID,
	}

	_, err = client.UpdateSiemSettings(ctx, updateSiemSettings)
	if err != nil {
		logger.Errorf("calling 'updateSiemSettings': %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceSiemSettingsRead(ctx, d, m)
}

func resourceSiemSettingsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := akamai.Meta(m)
	client := inst.Client(meta)
	logger := meta.Log("APPSEC", "resourceSiemSettingsDelete")
	logger.Debugf("in resourceSiemSettingsDelete")

	configID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	version, err := getModifiableConfigVersion(ctx, configID, "siemSetting", m)
	if err != nil {
		return diag.FromErr(err)
	}

	removeSiemSettings := appsec.RemoveSiemSettingsRequest{
		ConfigID:   configID,
		Version:    version,
		EnableSiem: false,
	}

	_, err = client.RemoveSiemSettings(ctx, removeSiemSettings)
	if err != nil {
		logger.Errorf("calling 'updateSiemSettings': %s", err.Error())
		return diag.FromErr(err)
	}
	return nil
}
