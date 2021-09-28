conda activate $CONDA_ENV
inv -e deps --verbose --dep-vendor-only
inv agent.version --major-version $MAJOR_VERSION -u > version.txt
cd $GOPATH/pkg && tar czf $CI_PROJECT_DIR/go-pkg.tar.gz .
cd $GOPATH/bin && tar czf $CI_PROJECT_DIR/go-bin.tar.gz .
cd $CI_PROJECT_DIR/vendor && tar czf $CI_PROJECT_DIR/vendor.tar.gz .
