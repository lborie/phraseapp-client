if [[ -z $GIT_COMMIT ]]; then
  GIT_COMMIT=$(git rev-parse HEAD)
fi

BUILT_AT="$(TZ=UTZ date +"%Y-%m-%dT%H:%M:%SZ")"
VERSION=$(cat VERSION)
CHANGES=$(if git status --porcelain > /dev/null 2>&1; then echo true; else false; fi)

if [[ $CHANGES == "true" ]]; then
	VERSION="${VERSION}-changes"
fi

build=$(cat <<EOF
{
  "built_at":      "$BUILT_AT",
  "revision":      "$GIT_COMMIT",
  "changes":       ${CHANGES},
  "hostname":      "$(hostname)",
  "user":          "$(whoami)",
  "version":	   "${VERSION}",
  "docs_revision": "1234"
}
EOF
)

echo $build | jq -c -r .
