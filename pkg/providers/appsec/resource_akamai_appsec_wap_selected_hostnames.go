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
func resourceWAPSelectedHostnames() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWAPSelectedHostnamesCreate,
		ReadContext:   resourceWAPSelectedHostnamesRead,
		UpdateContext: resourceWAPSelectedHostnamesUpdate,
		DeleteContext: resourceWAPSelectedHostnamesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: customdiff.All(
			VerifyIDUnchanged,
			verifyHostNotInBothLists,
		),
		Schema: map[string]*schema.Schema{
			"config_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"security_policy_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"protected_hosts": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"evaluated_hosts": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceWAPSelectedHostnamesCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := akamai.Meta(m)
	client := inst.Client(meta)
	logger := meta.Log("APPSEC", "resourceWAPSelectedHostnamesCreate")
	logger.Debugf("in resourceWAPSelectedHostnamesCreate")

	configID, err := tools.GetIntValue("config_id", d)
	if err != nil {
		return diag.FromErr(err)
	}
	securityPolicyID, err := tools.GetStringValue("security_policy_id", d)
	if err != nil {
		return diag.FromErr(err)
	}
	protectedHosts, err := tools.GetSetValue("protected_hosts", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}
	evaluatedHosts, err := tools.GetSetValue("evaluated_hosts", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}

	// convert to lists of strings
	var protectedHostnames, evalHostnames []string
	if (*protectedHosts).Len() > 0 {
		protectedHostnames = tools.SetToStringSlice(protectedHosts)
	} else {
		protectedHostnames = make([]string, 0)
	}
	if (*evaluatedHosts).Len() > 0 {
		evalHostnames = tools.SetToStringSlice(evaluatedHosts)
	} else {
		evalHostnames = make([]string, 0)
	}

	version, err := getModifiableConfigVersion(ctx, configID, "wapSelectedHostnames", m)
	if err != nil {
		return diag.FromErr(err)
	}
	updateWAPSelectedHostnames := appsec.UpdateWAPSelectedHostnamesRequest{
		ConfigID:         configID,
		Version:          version,
		SecurityPolicyID: securityPolicyID,
		ProtectedHosts:   protectedHostnames,
		EvaluatedHosts:   evalHostnames,
	}

	_, err = client.UpdateWAPSelectedHostnames(ctx, updateWAPSelectedHostnames)
	if err != nil {
		logger.Errorf("calling 'UpdateWAPSelectedHostnames': %s", err.Error())
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}

	if err := d.Set("config_id", configID); err != nil {
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}
	if err := d.Set("security_policy_id", securityPolicyID); err != nil {
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}
	if err := d.Set("protected_hosts", protectedHosts); err != nil {
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}
	if err := d.Set("evaluated_hosts", evaluatedHosts); err != nil {
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}

	d.SetId(fmt.Sprintf("%d:%s", configID, securityPolicyID))

	return resourceWAPSelectedHostnamesRead(ctx, d, m)
}

func resourceWAPSelectedHostnamesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := akamai.Meta(m)
	client := inst.Client(meta)
	logger := meta.Log("APPSEC", "resourceWAPSelectedHostnamesRead")
	logger.Debugf("in resourceWAPSelectedHostnamesRead")

	iDParts, err := splitID(d.Id(), 2, "configID:policyID")
	if err != nil {
		return diag.FromErr(err)
	}
	configID, err := strconv.Atoi(iDParts[0])
	if err != nil {
		return diag.FromErr(err)
	}
	version, err := getLatestConfigVersion(ctx, configID, m)
	if err != nil {
		return diag.FromErr(err)
	}
	securityPolicyID := iDParts[1]

	getWAPSelectedHostnamesRequest := appsec.GetWAPSelectedHostnamesRequest{
		ConfigID:         configID,
		Version:          version,
		SecurityPolicyID: securityPolicyID,
	}

	wapSelectedHostnames, err := client.GetWAPSelectedHostnames(ctx, getWAPSelectedHostnamesRequest)
	if err != nil {
		logger.Errorf("calling 'getWAPSelectedHostnames': %s", err.Error())
		return diag.FromErr(err)
	}

	if err := d.Set("config_id", configID); err != nil {
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}
	if err := d.Set("security_policy_id", securityPolicyID); err != nil {
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}
	if err := d.Set("protected_hosts", wapSelectedHostnames.ProtectedHosts); err != nil {
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}
	if err := d.Set("evaluated_hosts", wapSelectedHostnames.EvaluatedHosts); err != nil {
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}

	return nil
}

func resourceWAPSelectedHostnamesUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := akamai.Meta(m)
	client := inst.Client(meta)
	logger := meta.Log("APPSEC", "resourceWAPSelectedHostnamesUpdate")
	logger.Debugf("in resourceWAPSelectedHostnamesUpdate")

	configID, err := tools.GetIntValue("config_id", d)
	if err != nil {
		return diag.FromErr(err)
	}
	securityPolicyID, err := tools.GetStringValue("security_policy_id", d)
	if err != nil {
		return diag.FromErr(err)
	}
	protectedHosts, err := tools.GetSetValue("protected_hosts", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}
	evaluatedHosts, err := tools.GetSetValue("evaluated_hosts", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}

	// convert to lists of strings
	var protectedHostnames, evalHostnames []string
	if (*protectedHosts).Len() > 0 {
		protectedHostnames = tools.SetToStringSlice(protectedHosts)
	} else {
		protectedHostnames = make([]string, 0)
	}
	if (*evaluatedHosts).Len() > 0 {
		evalHostnames = tools.SetToStringSlice(evaluatedHosts)
	} else {
		evalHostnames = make([]string, 0)
	}

	version, err := getLatestConfigVersion(ctx, configID, m)
	if err != nil {
		return diag.FromErr(err)
	}
	updateWAPSelectedHostnames := appsec.UpdateWAPSelectedHostnamesRequest{
		ConfigID:         configID,
		Version:          version,
		SecurityPolicyID: securityPolicyID,
		ProtectedHosts:   protectedHostnames,
		EvaluatedHosts:   evalHostnames,
	}

	_, err = client.UpdateWAPSelectedHostnames(ctx, updateWAPSelectedHostnames)
	if err != nil {
		logger.Errorf("calling 'UpdateWAPSelectedHostnames': %s", err.Error())
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}

	if err := d.Set("config_id", configID); err != nil {
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}
	if err := d.Set("security_policy_id", securityPolicyID); err != nil {
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}
	if err := d.Set("protected_hosts", protectedHosts); err != nil {
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}
	if err := d.Set("evaluated_hosts", evaluatedHosts); err != nil {
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}

	return resourceWAPSelectedHostnamesRead(ctx, d, m)
}

func resourceWAPSelectedHostnamesDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return schema.NoopContext(ctx, d, m)
}

func verifyHostNotInBothLists(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	protectedHostsSet, err := tools.GetSetValue("protected_hosts", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return err
	}
	evaluatedHostsSet, err := tools.GetSetValue("evaluated_hosts", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return err
	}

	if protectedHostsSet.Len() == 0 && evaluatedHostsSet.Len() == 0 {
		return fmt.Errorf("protected_hostnames and evaluated_hostnames cannot both be empty")
	}

	if protectedHostsSet.Len() > 0 && evaluatedHostsSet.Len() > 0 {
		for _, h := range protectedHostsSet.List() {
			if evaluatedHostsSet.Contains(h) {
				return fmt.Errorf("host %s cannot be in both protected_hosts and evaluated_hosts lists", h)
			}
		}
		for _, h := range evaluatedHostsSet.List() {
			if protectedHostsSet.Contains(h) {
				return fmt.Errorf("host %s cannot be in both protected_hosts and evaluated_hosts lists", h)
			}
		}
	}

	return nil
}
