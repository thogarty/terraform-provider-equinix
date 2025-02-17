package equinix

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	equinix_errors "github.com/equinix/terraform-provider-equinix/internal/errors"
	equinix_schema "github.com/equinix/terraform-provider-equinix/internal/schema"

	"github.com/equinix/terraform-provider-equinix/internal/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	v4 "github.com/equinix-labs/fabric-go/fabric/v4"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceFabricRoutingProtocol() *schema.Resource {
	return &schema.Resource{
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(6 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(6 * time.Minute),
			Read:   schema.DefaultTimeout(6 * time.Minute),
		},
		ReadContext:   resourceFabricRoutingProtocolRead,
		CreateContext: resourceFabricRoutingProtocolCreate,
		UpdateContext: resourceFabricRoutingProtocolUpdate,
		DeleteContext: resourceFabricRoutingProtocolDelete,
		Importer: &schema.ResourceImporter{
			// Custom state context function, to parse import argument as  connection_uuid/rp_uuid
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				parts := strings.SplitN(d.Id(), "/", 2)
				if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
					return nil, fmt.Errorf("unexpected format of ID (%s), expected <conn-uuid>/<rp-uuid>", d.Id())
				}
				connectionUuid, uuid := parts[0], parts[1]
				// set set connection uuid and rp uuid as overall id of resource
				_ = d.Set("connection_uuid", connectionUuid)
				d.SetId(uuid)
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: createFabricRoutingProtocolResourceSchema(),

		Description: "Fabric V4 API compatible resource allows creation and management of Equinix Fabric connection\n\n~> **Note** Equinix Fabric v4 resources and datasources are currently in Beta. The interfaces related to `equinix_fabric_` resources and datasources may change ahead of general availability. Please, do not hesitate to report any problems that you experience by opening a new [issue](https://github.com/equinix/terraform-provider-equinix/issues/new?template=bug.md)",
	}
}

func resourceFabricRoutingProtocolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*config.Config).FabricClient
	ctx = context.WithValue(ctx, v4.ContextAccessToken, meta.(*config.Config).FabricAuthToken)
	log.Printf("[WARN] Routing Protocol Connection uuid: %s", d.Get("connection_uuid").(string))
	fabricRoutingProtocol, _, err := client.RoutingProtocolsApi.GetConnectionRoutingProtocolByUuid(ctx, d.Id(), d.Get("connection_uuid").(string))
	if err != nil {
		log.Printf("[WARN] Routing Protocol %s not found , error %s", d.Id(), err)
		if !strings.Contains(err.Error(), "500") {
			d.SetId("")
		}
		return diag.FromErr(equinix_errors.FormatFabricError(err))
	}
	switch fabricRoutingProtocol.Type_ {
	case "BGP":
		d.SetId(fabricRoutingProtocol.RoutingProtocolBgpData.Uuid)
	case "DIRECT":
		d.SetId(fabricRoutingProtocol.RoutingProtocolDirectData.Uuid)
	}

	return setFabricRoutingProtocolMap(d, fabricRoutingProtocol)
}

func resourceFabricRoutingProtocolCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*config.Config).FabricClient
	ctx = context.WithValue(ctx, v4.ContextAccessToken, meta.(*config.Config).FabricAuthToken)
	schemaBgpIpv4 := d.Get("bgp_ipv4").(*schema.Set).List()
	bgpIpv4 := routingProtocolBgpIpv4ToFabric(schemaBgpIpv4)
	schemaBgpIpv6 := d.Get("bgp_ipv6").(*schema.Set).List()
	bgpIpv6 := routingProtocolBgpIpv6ToFabric(schemaBgpIpv6)
	schemaDirectIpv4 := d.Get("direct_ipv4").(*schema.Set).List()
	directIpv4 := routingProtocolDirectIpv4ToFabric(schemaDirectIpv4)
	schemaDirectIpv6 := d.Get("direct_ipv6").(*schema.Set).List()
	directIpv6 := routingProtocolDirectIpv6ToFabric(schemaDirectIpv6)
	schemaBfd := d.Get("bfd").(*schema.Set).List()
	bfd := routingProtocolBfdToFabric(schemaBfd)
	bgpAuthKey := d.Get("bgp_auth_key")
	if bgpAuthKey == nil {
		bgpAuthKey = ""
	}

	createRequest := v4.RoutingProtocolBase{}
	if d.Get("type").(string) == "BGP" {
		createRequest = v4.RoutingProtocolBase{
			Type_: d.Get("type").(string),
			OneOfRoutingProtocolBase: v4.OneOfRoutingProtocolBase{
				RoutingProtocolBgpType: v4.RoutingProtocolBgpType{
					Type_:       d.Get("type").(string),
					Name:        d.Get("name").(string),
					BgpIpv4:     &bgpIpv4,
					BgpIpv6:     &bgpIpv6,
					CustomerAsn: int64(d.Get("customer_asn").(int)),
					EquinixAsn:  int64(d.Get("equinix_asn").(int)),
					BgpAuthKey:  bgpAuthKey.(string),
					Bfd:         &bfd,
				},
			},
		}
		if bgpIpv4.CustomerPeerIp == "" {
			createRequest.BgpIpv4 = nil
		}
		if bgpIpv6.CustomerPeerIp == "" {
			createRequest.BgpIpv6 = nil
		}
		if bfd.Enabled == false {
			createRequest.Bfd = nil
		}
	}
	if d.Get("type").(string) == "DIRECT" {
		createRequest = v4.RoutingProtocolBase{
			Type_: d.Get("type").(string),
			OneOfRoutingProtocolBase: v4.OneOfRoutingProtocolBase{
				RoutingProtocolDirectType: v4.RoutingProtocolDirectType{
					Type_:      d.Get("type").(string),
					Name:       d.Get("name").(string),
					DirectIpv4: &directIpv4,
					DirectIpv6: &directIpv6,
				},
			},
		}
		if directIpv4.EquinixIfaceIp == "" {
			createRequest.DirectIpv4 = nil
		}
		if directIpv6.EquinixIfaceIp == "" {
			createRequest.DirectIpv6 = nil
		}
	}
	fabricRoutingProtocol, _, err := client.RoutingProtocolsApi.CreateConnectionRoutingProtocol(ctx, createRequest, d.Get("connection_uuid").(string))
	if err != nil {
		return diag.FromErr(equinix_errors.FormatFabricError(err))
	}

	switch fabricRoutingProtocol.Type_ {
	case "BGP":
		d.SetId(fabricRoutingProtocol.RoutingProtocolBgpData.Uuid)
	case "DIRECT":
		d.SetId(fabricRoutingProtocol.RoutingProtocolDirectData.Uuid)
	}

	if _, err = waitUntilRoutingProtocolIsProvisioned(d.Id(), d.Get("connection_uuid").(string), meta, ctx); err != nil {
		return diag.Errorf("error waiting for RP (%s) to be created: %s", d.Id(), err)
	}

	return resourceFabricRoutingProtocolRead(ctx, d, meta)
}

func resourceFabricRoutingProtocolUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*config.Config).FabricClient
	ctx = context.WithValue(ctx, v4.ContextAccessToken, meta.(*config.Config).FabricAuthToken)

	schemaBgpIpv4 := d.Get("bgp_ipv4").(*schema.Set).List()
	bgpIpv4 := routingProtocolBgpIpv4ToFabric(schemaBgpIpv4)
	schemaBgpIpv6 := d.Get("bgp_ipv6").(*schema.Set).List()
	bgpIpv6 := routingProtocolBgpIpv6ToFabric(schemaBgpIpv6)
	schemaDirectIpv4 := d.Get("direct_ipv4").(*schema.Set).List()
	directIpv4 := routingProtocolDirectIpv4ToFabric(schemaDirectIpv4)
	schemaDirectIpv6 := d.Get("direct_ipv6").(*schema.Set).List()
	directIpv6 := routingProtocolDirectIpv6ToFabric(schemaDirectIpv6)
	schemaBfd := d.Get("bfd").(*schema.Set).List()
	bfd := routingProtocolBfdToFabric(schemaBfd)
	bgpAuthKey := d.Get("bgp_auth_key")
	if bgpAuthKey == nil {
		bgpAuthKey = ""
	}

	updateRequest := v4.RoutingProtocolBase{}
	if d.Get("type").(string) == "BGP" {
		updateRequest = v4.RoutingProtocolBase{
			Type_: d.Get("type").(string),
			OneOfRoutingProtocolBase: v4.OneOfRoutingProtocolBase{
				RoutingProtocolBgpType: v4.RoutingProtocolBgpType{
					Type_:       d.Get("type").(string),
					Name:        d.Get("name").(string),
					BgpIpv4:     &bgpIpv4,
					BgpIpv6:     &bgpIpv6,
					CustomerAsn: int64(d.Get("customer_asn").(int)),
					EquinixAsn:  int64(d.Get("equinix_asn").(int)),
					BgpAuthKey:  bgpAuthKey.(string),
					Bfd:         &bfd,
				},
			},
		}
		if bgpIpv4.CustomerPeerIp == "" {
			updateRequest.BgpIpv4 = nil
		}
		if bgpIpv6.CustomerPeerIp == "" {
			updateRequest.BgpIpv6 = nil
		}
	}
	if d.Get("type").(string) == "DIRECT" {
		updateRequest = v4.RoutingProtocolBase{
			Type_: d.Get("type").(string),
			OneOfRoutingProtocolBase: v4.OneOfRoutingProtocolBase{
				RoutingProtocolDirectType: v4.RoutingProtocolDirectType{
					Type_:      d.Get("type").(string),
					Name:       d.Get("name").(string),
					DirectIpv4: &directIpv4,
					DirectIpv6: &directIpv6,
				},
			},
		}
		if directIpv4.EquinixIfaceIp == "" {
			updateRequest.DirectIpv4 = nil
		}
		if directIpv6.EquinixIfaceIp == "" {
			updateRequest.DirectIpv6 = nil
		}
	}

	updatedRpResp, _, err := client.RoutingProtocolsApi.ReplaceConnectionRoutingProtocolByUuid(ctx, updateRequest, d.Id(), d.Get("connection_uuid").(string))
	if err != nil {
		return diag.FromErr(equinix_errors.FormatFabricError(err))
	}

	var changeUuid string
	switch updatedRpResp.Type_ {
	case "BGP":
		changeUuid = updatedRpResp.RoutingProtocolBgpData.Change.Uuid
		d.SetId(updatedRpResp.RoutingProtocolBgpData.Uuid)
	case "DIRECT":
		changeUuid = updatedRpResp.RoutingProtocolDirectData.Change.Uuid
		d.SetId(updatedRpResp.RoutingProtocolDirectData.Uuid)
	}
	_, err = waitForRoutingProtocolUpdateCompletion(changeUuid, d.Id(), d.Get("connection_uuid").(string), meta, ctx)
	if err != nil {
		if !strings.Contains(err.Error(), "500") {
			d.SetId("")
		}
		return diag.FromErr(fmt.Errorf("timeout updating routing protocol: %v", err))
	}
	updatedProvisionedRpResp, err := waitUntilRoutingProtocolIsProvisioned(d.Id(), d.Get("connection_uuid").(string), meta, ctx)
	if err != nil {
		return diag.Errorf("error waiting for RP (%s) to be replace updated: %s", d.Id(), err)
	}

	return setFabricRoutingProtocolMap(d, updatedProvisionedRpResp)
}

func resourceFabricRoutingProtocolDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}
	client := meta.(*config.Config).FabricClient
	ctx = context.WithValue(ctx, v4.ContextAccessToken, meta.(*config.Config).FabricAuthToken)
	_, _, err := client.RoutingProtocolsApi.DeleteConnectionRoutingProtocolByUuid(ctx, d.Id(), d.Get("connection_uuid").(string))
	if err != nil {
		errors, ok := err.(v4.GenericSwaggerError).Model().([]v4.ModelError)
		if ok {
			// EQ-3142509 = Connection already deleted
			if equinix_errors.HasModelErrorCode(errors, "EQ-3142509") {
				return diags
			}
		}
		return diag.FromErr(equinix_errors.FormatFabricError(err))
	}

	err = waitUntilRoutingProtocolIsDeprovisioned(d.Id(), d.Get("connection_uuid").(string), meta, ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("API call failed while waiting for resource deletion. Error %v", err))
	}

	return diags
}

func setFabricRoutingProtocolMap(d *schema.ResourceData, rp v4.RoutingProtocolData) diag.Diagnostics {
	diags := diag.Diagnostics{}

	err := error(nil)
	if rp.Type_ == "BGP" {
		err = equinix_schema.SetMap(d, map[string]interface{}{
			"name":         rp.RoutingProtocolBgpData.Name,
			"href":         rp.RoutingProtocolBgpData.Href,
			"type":         rp.RoutingProtocolBgpData.Type_,
			"state":        rp.RoutingProtocolBgpData.State,
			"operation":    routingProtocolOperationToTerra(rp.RoutingProtocolBgpData.Operation),
			"bgp_ipv4":     routingProtocolBgpConnectionIpv4ToTerra(rp.BgpIpv4),
			"bgp_ipv6":     routingProtocolBgpConnectionIpv6ToTerra(rp.BgpIpv6),
			"customer_asn": rp.CustomerAsn,
			"equinix_asn":  rp.EquinixAsn,
			"bfd":          routingProtocolBfdToTerra(rp.Bfd),
			"bgp_auth_key": rp.BgpAuthKey,
			"change":       routingProtocolChangeToTerra(rp.RoutingProtocolBgpData.Change),
			"change_log":   changeLogToTerra(rp.RoutingProtocolBgpData.Changelog),
		})
	} else if rp.Type_ == "DIRECT" {
		err = equinix_schema.SetMap(d, map[string]interface{}{
			"name":        rp.RoutingProtocolDirectData.Name,
			"href":        rp.RoutingProtocolDirectData.Href,
			"type":        rp.RoutingProtocolDirectData.Type_,
			"state":       rp.RoutingProtocolDirectData.State,
			"operation":   routingProtocolOperationToTerra(rp.RoutingProtocolDirectData.Operation),
			"direct_ipv4": routingProtocolDirectConnectionIpv4ToTerra(rp.DirectIpv4),
			"direct_ipv6": routingProtocolDirectConnectionIpv6ToTerra(rp.DirectIpv6),
			"change":      routingProtocolChangeToTerra(rp.RoutingProtocolDirectData.Change),
			"change_log":  changeLogToTerra(rp.RoutingProtocolDirectData.Changelog),
		})
	}
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func waitUntilRoutingProtocolIsProvisioned(uuid string, connUuid string, meta interface{}, ctx context.Context) (v4.RoutingProtocolData, error) {
	log.Printf("Waiting for routing protocol to be provisioned, uuid %s", uuid)
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			string(v4.PROVISIONING_ConnectionState),
			string(v4.REPROVISIONING_ConnectionState),
		},
		Target: []string{
			string(v4.PROVISIONED_ConnectionState),
		},
		Refresh: func() (interface{}, string, error) {
			client := meta.(*config.Config).FabricClient
			dbConn, _, err := client.RoutingProtocolsApi.GetConnectionRoutingProtocolByUuid(ctx, uuid, connUuid)
			if err != nil {
				return "", "", equinix_errors.FormatFabricError(err)
			}
			var state string
			if dbConn.Type_ == "BGP" {
				state = dbConn.RoutingProtocolBgpData.State
			} else if dbConn.Type_ == "DIRECT" {
				state = dbConn.RoutingProtocolDirectData.State
			}
			return dbConn, state, nil
		},
		Timeout:    5 * time.Minute,
		Delay:      30 * time.Second,
		MinTimeout: 30 * time.Second,
	}

	inter, err := stateConf.WaitForStateContext(ctx)
	dbConn := v4.RoutingProtocolData{}

	if err == nil {
		dbConn = inter.(v4.RoutingProtocolData)
	}

	return dbConn, err
}

func waitUntilRoutingProtocolIsDeprovisioned(uuid string, connUuid string, meta interface{}, ctx context.Context) error {
	log.Printf("Waiting for routing protocol to be deprovisioned, uuid %s", uuid)

	/* check if resource is not found */
	stateConf := &retry.StateChangeConf{
		Target: []string{
			strconv.Itoa(404),
		},
		Refresh: func() (interface{}, string, error) {
			client := meta.(*config.Config).FabricClient
			dbConn, resp, _ := client.RoutingProtocolsApi.GetConnectionRoutingProtocolByUuid(ctx, uuid, connUuid)
			// fixme: check for error code instead?
			// ignore error for Target
			return dbConn, strconv.Itoa(resp.StatusCode), nil

		},
		Timeout:    5 * time.Minute,
		Delay:      30 * time.Second,
		MinTimeout: 30 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func waitForRoutingProtocolUpdateCompletion(rpChangeUuid string, uuid string, connUuid string, meta interface{}, ctx context.Context) (v4.RoutingProtocolChangeData, error) {
	log.Printf("Waiting for routing protocol update to complete, uuid %s", uuid)
	stateConf := &retry.StateChangeConf{
		Target: []string{"COMPLETED"},
		Refresh: func() (interface{}, string, error) {
			client := meta.(*config.Config).FabricClient
			dbConn, _, err := client.RoutingProtocolsApi.GetConnectionRoutingProtocolsChangeByUuid(ctx, connUuid, uuid, rpChangeUuid)
			if err != nil {
				return "", "", equinix_errors.FormatFabricError(err)
			}
			updatableState := ""
			if dbConn.Status == "COMPLETED" {
				updatableState = dbConn.Status
			}
			return dbConn, updatableState, nil
		},
		Timeout:    2 * time.Minute,
		Delay:      30 * time.Second,
		MinTimeout: 30 * time.Second,
	}

	inter, err := stateConf.WaitForStateContext(ctx)
	dbConn := v4.RoutingProtocolChangeData{}

	if err == nil {
		dbConn = inter.(v4.RoutingProtocolChangeData)
	}
	return dbConn, err
}
