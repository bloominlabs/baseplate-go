#!/usr/bin/env bash
DEPENDABOT_PATH="./.github/dependabot.yml"
cp ./scripts/dependabot.base.yml $DEPENDABOT_PATH

for d in `find . -type f -name "go.mod" -not -path '*/deploy/*' -exec dirname {} \; | sort | egrep '^./'`; do
  echo "updating " $d
  echo "  - package-ecosystem: gomod" >> $DEPENDABOT_PATH
  echo "    directory: $d" >> $DEPENDABOT_PATH
  echo "    schedule:" >> $DEPENDABOT_PATH
  echo "      interval: weekly" >> $DEPENDABOT_PATH
  echo "      day: tuesday" >> $DEPENDABOT_PATH
done
