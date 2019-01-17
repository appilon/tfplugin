# TROUBLSHOOTING

## Upgrading to terraform 0.12
I have noticed for some providers this leads to a compilation error in a transitive dependency? It might be worth specifying `-u` which will upgrade transitive dependencies to their latest `MINOR` release, this might solve the issue however in my specific experiences it fails without a helpful error message.

```
$ GO111MODULE=on go get -u github.com/hashicorp/terraform@pluginsdk-v0.12-early2
```

UPDATE
Digging deeper there appears to be a problem with a transitive dependency `cloud.google.com/go`. This is relied on by many projects, notably grpc which itself is relied on by a lot of this. you can inspect a providers graph with

```
$ GO111MODULE=on go mod graph
```

If you encounter an error when upgrading to the new sdk regarding that dependency, try this.
```
$ GO111MODULE=on go get -u cloud.google.com/go@master
$ GO111MODULE=on go get github.com/hashicorp/terraform@pluginsdk-v0.12-early2
$ GO111MODULE=on go mod tidy
$ GO111MODULE=on go mod vendor
```

## HCL upgrades
These notes are very rough but they describe common issues and the fixes for them.

1)
```
"id": {
                Type:     schema.TypeString,
                Computed: true,
            },
```
This is an automatic attribute, do not declare it or InternalValidate() will fail

2)
The most common issue will be nested types such as Map and List now requiring an equals assignment, and nested types that are resources now needing to be a true nested blocks (no equals). Martin's words explain it best:

>From a provider developer's perspective, that should be described as: any collection that has its `Elem` set to a `schema.Schema` is an "attribute" which must be set with the equals sign, while any where `Elem` is set to a `schema.Resource` is a "nested block" which must _not_ have the equals sign

3)
```
resource "grafana_alert_notification" "test" {
    type = "email"
    name = "terraform-acc-test"
    settings {
			"addresses" = "foo@bar.test"
			"uploadImage" = "false"
			"autoResolve" = "true"
		}
}
```
In this situation we first encountered a parsing error claiming attributes cannot be quoted. This is because Terraform hasn't processed the schema yet and doesn't realize the real error is #2 ^. If you do remove the quotes (which are allowed for TypeMap), you will get the same error message as #2.

4)
Many of our test cases will check that a list is empty. In 0.12 if the config had no attribute at all it will not be in the state, therefore we _had_ failures as a result of the key/path not even existing. This has since been shimmed, however it might be good to switch these checks to use `TestCheckNoResourceAttr`
```
{
				Config: testAccOrganizationConfig_usersUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationCheckExists("grafana_organization.test", &org),
					resource.TestCheckResourceAttr(
						"grafana_organization.test", "name", "terraform-acc-test",
					),
					resource.TestCheckResourceAttr(
						"grafana_organization.test", "admins.#", "0",
					),
					resource.TestCheckResourceAttr(
						"grafana_organization.test", "editors.#", "1",
					),
					resource.TestCheckResourceAttr(
						"grafana_organization.test", "editors.0", "john.doe@example.com",
					),
					resource.TestCheckResourceAttr(
						"grafana_organization.test", "viewers.#", "0",
					),
				),
			},
```