# Manage example account.
resource "zesty_account" "example" {
  account = {
    id             = "123456789012"
    cloud_provider = "AWS"
    role_arn       = "arn:aws:iam::123456789012:role/ZestyIamRole"
    external_id    = "f1f0a7f7-a523-4197-9e19-ffd205a5bc20"
    products = {
      kompass = {
        active = true
      }
    }
  }
}
