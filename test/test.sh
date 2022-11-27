# This script will run a quick demo of the merge driver.
# It will cause a merge conflict in a 'my-file.mrg' file.
# -------------------------------------------------------

# Run the mergetool-setup.sh script to configure the merge driver
./mergetool-setup.sh

# Add the my-merge-tool.sh to PATH
PATH=$PATH:`pwd`

# Clean up any previous example runs
git checkout main
git branch -D demo-branch-1
git branch -D demo-branch-2

# Create 'resources.elements.json' on branch 1
git checkout -b demo-branch-1
cp ../testdata/resources.elements.A.json resources.elements.json
git add resources.elements.json
git commit -m"demo-branch-1: added resources.elements.json"

# Create 'resources.elements.json' on branch 2
git checkout main
git checkout -b demo-branch-2
cp ../testdata/resources.elements.B.json resources.elements.json
git add resources.elements.json
git commit -m"demo-branch-2: added resources.elements.json"

# Merge the two branches, causing a conflict
git merge -m"Merged in demo-branch-1" demo-branch-1

git checkout main