# On Leave Form Page

This directory contains the static HTML page for the on-leave form.

## Manual Deployment

To deploy this form to Cloudflare Pages, run the following command from this directory:

```bash
npx wrangler pages deploy . --project-name=onleave-form --commit-dirty=true
```

This will deploy the form page to Cloudflare Pages under the `onleave-form` project.
