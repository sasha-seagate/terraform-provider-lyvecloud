package lyvecloud

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourcePermissionV2() *schema.Resource {

	return &schema.Resource{
		Create: resourcePermissionV2Create,
		Read:   resourcePermissionV2Read,
		Update: resourcePermissionV2Update,
		Delete: resourcePermissionV2Delete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"name"},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true, // should be generated if empty
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true, // computed based on the chosen argument. all_buckets/prefix/buckets
			},
			"actions": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"all-operations",
					"read-only",
					"write-only",
				}, false),
			},
			"all_buckets": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"buckets", "buckets_prefix"},
			},
			"buckets_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"buckets", "all_buckets"},
			},
			"buckets": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ConflictsWith: []string{"buckets_prefix", "all_buckets"},
			},
			"ready_state": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourcePermissionV2Create(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(Account, meta.(Client)) {
		return fmt.Errorf("credentials for account api(client_id, client_secret) are missing")
	}

	conn := *meta.(Client).AccAPIV2Client

	name := NameWithSuffix(d.Get("name").(string), d.Get("name_prefix").(string))
	description := d.Get("description").(string)
	actions := d.Get("actions").(string)
	prefix := d.Get("buckets_prefix").(string)

	var permissionType string
	buckets := []string{}

	if _, ok := d.GetOk("all_buckets"); ok {
		permissionType = "all-buckets"
	} else if v, ok := d.GetOk("bucket_prefix"); ok {
		permissionType = "bucket-prefix"
		buckets = append(buckets, v.(string))
	} else if _, ok := d.GetOk("bucket_names"); ok {
		permissionType = "bucket-names"
		bucketsList := d.Get("buckets").([]interface{})
		for _, v := range bucketsList {
			buckets = append(buckets, v.(string))
		}
	}

	// create input for CreatePermissionV2
	createPermissinInput := PermissionV2{
		Name:        name,
		Description: description,
		Type:        permissionType,
		Actions:     actions,
		Prefix:      prefix,
		Buckets:     buckets,
	}

	resp, err := conn.CreatePermissionV2(&createPermissinInput)
	if err != nil {
		return fmt.Errorf("error creating permission: %w", err)
	}
	d.SetId(resp.ID)

	return resourcePermissionV2Read(d, meta)
}

func resourcePermissionV2Read(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(Account, meta.(Client)) {
		return fmt.Errorf("credentials for account api(client_id, client_secret) are missing")
	}

	conn := *meta.(Client).AccAPIV2Client

	permissionId := d.Id()

	resp, err := conn.GetPermissionV2(permissionId)
	if err != nil {
		return fmt.Errorf("error reading permission: %w", err)
		// try to remove it if error?
	}

	d.Set("id", resp.Id)
	d.Set("name", resp.Name)
	d.Set("description", resp.Description)
	d.Set("type", resp.Type)
	d.Set("actions", resp.Actions)
	d.Set("buckets_prefix", resp.Prefix)
	d.Set("bucket_names", resp.Buckets)
	d.Set("ready_state", resp.ReadyState)

	return nil
}

func resourcePermissionV2Update(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(Account, meta.(Client)) {
		return fmt.Errorf("credentials for account api(client_id, client_secret) are missing")
	}

	conn := *meta.(Client).AccAPIV2Client

	permissionId := d.Id()

	name := NameWithSuffix(d.Get("name").(string), d.Get("name_prefix").(string))
	description := d.Get("description").(string)
	actions := d.Get("actions").(string)
	prefix := d.Get("buckets_prefix").(string)

	var permissionType string
	buckets := []string{}

	if _, ok := d.GetOk("all_buckets"); ok {
		permissionType = "all-buckets"
	} else if v, ok := d.GetOk("buckets_prefix"); ok {
		permissionType = "bucket-prefix"
		buckets = append(buckets, v.(string))
	} else if _, ok := d.GetOk("buckets"); ok {
		permissionType = "bucket-names"
		bucketsList := d.Get("buckets").([]interface{})
		for _, v := range bucketsList {
			buckets = append(buckets, v.(string))
		}
	}

	updatePermissinInput := PermissionV2{
		Name:        name,
		Description: description,
		Type:        permissionType,
		Actions:     actions,
		Prefix:      prefix,
		Buckets:     buckets,
	}

	_, err := conn.UpdatePermissionV2(permissionId, &updatePermissinInput)
	if err != nil {
		return fmt.Errorf("error updating permission: %w", err)
	}

	return resourcePermissionV2Read(d, meta)
}

func resourcePermissionV2Delete(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(Account, meta.(Client)) {
		return fmt.Errorf("credentials for account api(client_id, client_secret) are missing")
	}

	conn := *meta.(Client).AccAPIV2Client

	_, err := conn.DeletePermissionV2(d.Id())
	if err != nil {
		return fmt.Errorf("error deleting permission: %w", err)
	}

	return nil
}