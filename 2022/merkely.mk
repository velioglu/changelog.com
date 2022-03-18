# 💻 https://github.com/merkely-development/cli
# 📦 https://github.com/merkely-development/cli/pkgs/container/merkely-cli/16382763?tag=v1.5.0
# 🔒 sha256:00bf2c59e691f95dba3d2c7918d32500d5c319158b01879ec8f888f660645e0c
#
# 🚧 TODO: How do I verify the image?
# Manually today, using the show from the ghcr.io package.
# cosign/sigstore is one of the potential future solutions.
#
# 👉 https://app.merkely.com/changelog/pipelines/
# merkely pipeline declare \
# 	--api-token ${MERKELY_API_TOKEN \
# 	--description yourPipelineDescription \
# 	--template artifact \
# 	--pipeline 2022-changelog-dagger \
# 	--owner changelog \
# 	--visibility public
#
# 💡 --template maps to evidence
# 💡 e.g. https://app.merkely.com/cyber-dojo/pipelines/differ/artifacts/83c8b5b2a65b7381a87eb43a92acddd2a1960bd8bc6164d0c38a5714d4675b7fhttps://app.merkely.com/cyber-dojo/pipelines/differ/artifacts/83c8b5b2a65b7381a87eb43a92acddd2a1960bd8bc6164d0c38a5714d4675b7f
#
# merkely pipeline artifact report creation thechangelog/changelog.com:master
# 	--api-token ${MERKELY_API_TOKEN} \
# 	--build-url ${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID} \
# 	--commit-url ${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/commit/${GITHUB_SHA} \
# 	--compliant \
# 	--description "https://hub.docker.com/layers/thechangelog/changelog.com/master/images/sha256-2e36739ea2acb215346e2b488f15824329b9c7c3e23563ca0df04fa8ab9a900f?context=explore" \
# 	--git-commit ${GITHUB_SHA} \
# 	--owner changelog \
# 	--pipeline 2022-changelog-dagger \
# 	--sha256 explicit-image-digest
#
# 💡 If there is an issue with api.merkely.com, use --dry-run
#
# 📚 https://app.merkely.com/cyber-dojo/environments/
# 📚 https://cyber-dojo.org/creator/home
# 📚 https://www.merkely.com/blog/my-first-week-at-merkely/
# 📚 https://www.merkely.com/blog/8-reasons-why-we-do-ensemble-programming/
