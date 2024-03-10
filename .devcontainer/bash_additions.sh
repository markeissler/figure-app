function parse_git_dirty {
  _str=$(git status --porcelain --untracked 2> /dev/null | tail -n1);

  if [[ "${_str}" == "??"* ]]; then
    echo "+";
  elif [[ "${_str}" == " D "* ]]; then
    echo "-"
  elif [[ "${_str}" == " M "* ]]; then
    echo "*"
  fi
}
function parse_git_branch {
  git branch --no-color 2> /dev/null | sed -e '/^[^*]/d' -e "s/* \(.*\)/[\1$(parse_git_dirty)]/"
}
export PS1='\n\u@\h::\[\033[1;33m\]\w\[\033[0m\]$(parse_git_branch)\nprompt> '

# Add some aliases
# git oneline log support
alias git-lg0="git log --pretty=oneline --abbrev-commit"

# git commits in dev but not master
alias git-ready-dev="git log --oneline dev ^master"
alias git-ready-develop="git log --oneline develop ^master"

# git tags sorted for SemVer
alias git-tags="git tag -n99 | sort -k1 --version-sort"

# Pre-emptively run direnv allow
if command -v direnv &> /dev/null && test -f "$PWD/.envrc"; then
    direnv allow
fi
