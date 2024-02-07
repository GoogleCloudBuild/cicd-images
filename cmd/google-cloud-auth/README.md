# google-cloud-auth

A CLI command tool to generate authencitation files for gcloud or Cloud Client Library

## Get Credential File
Embeded the `google-cloud-auth` image to your specific Gitlab Pipeline/Components and run `google-cloud-auth generate-credentials ...` commands 
to generate Workload Identity Federation credential file to authenticate requests to GCP via [gcloud][gcloud] or [Google Cloud Client Libraries][cloud-client-lib].

To authenticate with [gcloud][gcloud], you need to set `GCLOUDSDK_AUTH_CREDENTIAL_FILE_OVERRIDE` to the generated credential file path.

To authenticate with [Google Client Library][cloud-client-lib], you need to set [`GOOGLE_APPLICATION_CREDENTIALS`][lib-auth] to the generated credential file path.

## Inputs
-   `oidc-jwt_env_var`: (Required) The Env Var (without "$") containing full OIDC JWT provided by Gitlab, can be found as `id_tokens.GCP_OIDC_JWT` in the
     Gitlab CI/CD config.

    ```text
    id_tokens:
        GCP_OIDC_JWT:
        aud: ...
    ```
-   `workload-identity-provider`: (Required) The full identifier of the Workload
    Identity Provider, including the project number, pool name, and provider
    name. If provided, this must be the full identifier which includes all
    parts:

    ```text
    //iam.googleapis.com/projects/<project-number>/locations/global/workloadIdentityPools/<pool-id>/providers/<provider-id>
    ```

-   `service-account`: (Optional) Email address or unique identifier of the
    Google Cloud service account for which to impersonate and generate
    credentials. For example:

    ```text
    my-service-account@my-project.iam.gserviceaccount.com
    ```

    Without this input, the Gitlab Components using this binary will use Direct Workload Identity
    Federation. If this input is provided, the Gitlab Components will use
    Workload Identity Federation through a Service Account.

-   `credentials-json-output-path`: (Optional) The full file path of the output credentials json, default to `/tmp/oidc-jwt.json`.


[secure-file]: https://docs.gitlab.com/ee/ci/secure_files/
[cloud-client-lib]: https://cloud.google.com/apis/docs/cloud-client-libraries
[gcloud]: https://cloud.google.com/sdk?hl=en
[cloud-deploy]: https://gitlab.com/quanzhang/my-component/-/blob/main/templates/cloud-deploy-img.yml?ref_type=6a8ad3e2697d1a01a9d27f092f05c5c3099ab405
[lib-auth]: https://cloud.google.com/docs/authentication/application-default-credentials
