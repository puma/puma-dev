go get -u github.com/mitchellh/gox
go get -u github.com/tcnksm/ghr

export OWNER="nonrational"
export REPO="puma-dev"
export RELEASE="0.13.nonrational"
export GITHUB_TOKEN="$GITHUB_API_TOKEN"

make release

git tag -f "v${RELEASE}"
git push nonrational "v${RELEASE}"

ghr -u $OWNER  -t $GITHUB_TOKEN -r $REPO  -n "v${RELEASE}" -delete -prerelease "v${RELEASE}" ./pkg/
