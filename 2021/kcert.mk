# Staging
ACME_DIRURL := https://acme-staging-v02.api.letsencrypt.org/directory
# Uncomment production when ready:
# ACME_DIRURL := https://acme-v02.api.letsencrypt.org/directory
lke-kcert: | lke-ctx $(ENVSUBST)
	export ACME__DIRURL=$(ACME_DIRURL) \
	export ACME__TERMSACCEPTED=true \
	export ACME__EMAIL=$(CERT_EMAIL) \
	; cat $(CURDIR)/manifests/kcert.yml \
	| $(ENVSUBST_SAFE) \
	| $(KUBECTL) $(K_CMD) --filename -
