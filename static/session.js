toastr.options = {
  timeOut: 100,
  closeDuration: 100
};
window.addEventListener('online', () => toastr.success('you are back online'));
window.addEventListener('offline', () => toastr.error('you are offline'));

api.defaultErrorHandler = res => toastr.error(res.error.error);

const ClickSound = {
  init(url) {
    try {
      this.sound = new Audio(url);
      this.sound.volume = 0.2;
    } catch (error) {}
  },

  play() {
    if (this.sound) {
      try {
        this.sound.pause();
        this.sound.currentTime = 0;
        this.sound.play();
      } catch (error) {}
    }
  }

};
ClickSound.init("/static/n.mp3");

class FibSeq {
  constructor(options = {}) {
    this.outOfBucketLimit = options.bucketThreshold || 3;
    this.seqSize = options.size || 13;
    this.generateFibSeq_();
  }

  getSequence() {
    return this.seqs_.slice(0);
  }

  isOutOfBucket(sortedArray) {
    var decision = {
      overLimit: false,
      reason: null
    }; // Bail if unique scores are more than "threshhold"

    var distinct = 0,
        i = 0,
        prev = sortedArray[0];

    for (; i < sortedArray.length; i++) {
      var current = sortedArray[i];
      if (prev !== current) distinct++;

      if (distinct >= this.outOfBucketLimit) {
        decision.overLimit = true;
        decision.reason = `Out of ${this.outOfBucketLimit} buckets`;
        return decision;
      }

      prev = current;
    } // Bail if distance between index distance btw the first and last scores 
    // are more than "threshhold"


    if (this.seqs_.length > 1) {
      var minIndex = this.seqs_.indexOf(sortedArray[0]);
      var maxIndex = this.seqs_.indexOf(sortedArray[sortedArray.length - 1]);

      if (maxIndex - minIndex >= this.outOfBucketLimit) {
        decision.overLimit = true;
        decision.reason = `Distance(Max,Min) > ${this.outOfBucketLimit}`;
        return decision;
      }
    }

    return decision;
  }

  closestFib(val) {
    for (var i = 0; i < this.seqs_.length; i++) {
      if (this.seqs_[i] === val) return this.seqs_[i];

      if (val < this.seqs_[i]) {
        // if val is greater than last element
        // then compute next fib number and check difference
        var rightDiff = this.seqs_[i] - val;

        if (0 < i) {
          var leftDiff = val - this.seqs_[i - 1];
          return leftDiff < rightDiff ? this.seqs_[i - 1] : this.seqs_[i];
        }

        return this.seqs_[i];
      }
    }

    return val;
  }

  generateFibSeq_() {
    var n = Math.min(parseInt(this.seqSize) || 0, 20);
    var i = 0;
    var seq = [];

    while (i < n) {
      if (!seq.length) {
        seq.push(0);
      } else if (seq.length == 1) {
        seq.push(1);
      } else {
        seq.push(seq[i - 2] + seq[i - 1]);
      }

      i++;
    }

    this.seqs_ = [...new Set(seq)];
  }

}

const Spinner = props => /*#__PURE__*/React.createElement("div", {
  className: "spinner-grow text-primary",
  role: "status"
}, /*#__PURE__*/React.createElement("span", {
  className: "sr-only"
}, "props.message"));

class SessionStart extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      voters: [],
      error: null,
      isLoading: true
    };
    this.handleOpen = this.handleOpen.bind(this);
    this.handleCheckChange = this.handleCheckChange.bind(this);
  }

  handleOpen(e) {
    this.props.onOpen(this.state.voters.filter(v => v.checked));
  }

  handleCheckChange(e) {
    const voter = this.state.voters.find(v => v.name == e.target.value);

    if (voter) {
      voter.checked = e.target.checked;
    }

    this.setState({
      voters: this.state.voters.map(v => v)
    });
  }

  componentDidMount() {
    api.listUsers(users => {
      this.setState({
        isLoading: false,
        voters: users.filter(u => u.role == "voter").map(v => Object.assign(v, {
          checked: true
        }))
      });
    }, error => {
      this.setState({
        error,
        isLoading: false
      });
    });
  }

  render() {
    let component;

    if (this.state.isLoading) {
      component = /*#__PURE__*/React.createElement(Spinner, {
        message: "Loading voters list..."
      });
    } else if (this.state.error) {
      component = /*#__PURE__*/React.createElement("div", {
        className: "alert alert-danger"
      }, "Failed to load users, try to refresh the page");
    } else if (!this.state.voters.length) {
      component = /*#__PURE__*/React.createElement("div", {
        className: "alert alert-warning"
      }, "There is no any voter in the team yet, ", /*#__PURE__*/React.createElement("a", {
        href: "/ui/users",
        className: "bold"
      }, "add voter"), ".", /*#__PURE__*/React.createElement("br", null));
    } else {
      component = /*#__PURE__*/React.createElement(React.Fragment, null, /*#__PURE__*/React.createElement("h6", null, "Select voters"), /*#__PURE__*/React.createElement("div", {
        id: "SessionVoters",
        className: "mt-3"
      }, " ", this.state.voters.map((v, i) => /*#__PURE__*/React.createElement("div", {
        className: "form-check mt-2",
        key: v + i
      }, /*#__PURE__*/React.createElement("input", {
        className: "form-check-input voter-checkbox",
        name: "voter",
        onChange: this.handleCheckChange,
        type: "checkbox",
        value: v.name,
        checked: v.checked
      }), /*#__PURE__*/React.createElement("label", {
        className: "form-check-label"
      }, v.name)))), /*#__PURE__*/React.createElement("button", {
        id: "OpenSessionBtn",
        onClick: this.handleOpen,
        className: "btn btn-sm btn-primary mt-3"
      }, "Open ", /*#__PURE__*/React.createElement("svg", {
        viewBox: "0 0 16 16",
        width: "16",
        height: "16",
        className: "ml-1"
      }, /*#__PURE__*/React.createElement("path", {
        fillRule: "evenodd",
        d: "M14.064 0a8.75 8.75 0 00-6.187 2.563l-.459.458c-.314.314-.616.641-.904.979H3.31a1.75 1.75 0 00-1.49.833L.11 7.607a.75.75 0 00.418 1.11l3.102.954c.037.051.079.1.124.145l2.429 2.428c.046.046.094.088.145.125l.954 3.102a.75.75 0 001.11.418l2.774-1.707a1.75 1.75 0 00.833-1.49V9.485c.338-.288.665-.59.979-.904l.458-.459A8.75 8.75 0 0016 1.936V1.75A1.75 1.75 0 0014.25 0h-.186zM10.5 10.625c-.088.06-.177.118-.266.175l-2.35 1.521.548 1.783 1.949-1.2a.25.25 0 00.119-.213v-2.066zM3.678 8.116L5.2 5.766c.058-.09.117-.178.176-.266H3.309a.25.25 0 00-.213.119l-1.2 1.95 1.782.547zm5.26-4.493A7.25 7.25 0 0114.063 1.5h.186a.25.25 0 01.25.25v.186a7.25 7.25 0 01-2.123 5.127l-.459.458a15.21 15.21 0 01-2.499 2.02l-2.317 1.5-2.143-2.143 1.5-2.317a15.25 15.25 0 012.02-2.5l.458-.458h.002zM12 5a1 1 0 11-2 0 1 1 0 012 0zm-8.44 9.56a1.5 1.5 0 10-2.12-2.12c-.734.73-1.047 2.332-1.15 3.003a.23.23 0 00.265.265c.671-.103 2.273-.416 3.005-1.148z"
      }))), /*#__PURE__*/React.createElement("div", {
        className: "mt-3 small"
      }, "Opening a session makes you leader ", /*#__PURE__*/React.createElement("span", {
        className: "leader-badge"
      }, "\u2605"), " of it. ", /*#__PURE__*/React.createElement("a", {
        href: "/ui/docs"
      }, "More")));
    }

    return /*#__PURE__*/React.createElement("div", null, component);
  }

}

const VotingHeader = props => /*#__PURE__*/React.createElement("div", {
  className: "header row"
}, /*#__PURE__*/React.createElement("div", {
  className: "col-9 col-md-5"
}, /*#__PURE__*/React.createElement("div", {
  className: "session-name"
}, props.name, props.leader && /*#__PURE__*/React.createElement("span", {
  className: "leader-badge ml-1"
}, "\u2605"))), /*#__PURE__*/React.createElement("div", {
  className: "col-md-4 d-none d-md-flex"
}, props.result && /*#__PURE__*/React.createElement("div", {
  className: "primary-result-ready"
}, /*#__PURE__*/React.createElement("span", {
  className: "agg-name"
}, props.result.primaryAggregate.name), /*#__PURE__*/React.createElement("span", {
  className: "agg-value"
}, props.result.primaryAggregate.value))), /*#__PURE__*/React.createElement("div", {
  className: "col-3 col-md-3 text-right"
}, /*#__PURE__*/React.createElement("button", {
  type: "button",
  title: "Reset",
  className: "btn btn-reset btn-sm btn-lemon btn-fixed-1",
  onClick: props.onReset
}, /*#__PURE__*/React.createElement("svg", {
  viewBox: "0 0 16 16",
  width: "16",
  height: "16"
}, /*#__PURE__*/React.createElement("path", {
  fillRule: "evenodd",
  d: "M8 2.5a5.487 5.487 0 00-4.131 1.869l1.204 1.204A.25.25 0 014.896 6H1.25A.25.25 0 011 5.75V2.104a.25.25 0 01.427-.177l1.38 1.38A7.001 7.001 0 0114.95 7.16a.75.75 0 11-1.49.178A5.501 5.501 0 008 2.5zM1.705 8.005a.75.75 0 01.834.656 5.501 5.501 0 009.592 2.97l-1.204-1.204a.25.25 0 01.177-.427h3.646a.25.25 0 01.25.25v3.646a.25.25 0 01-.427.177l-1.38-1.38A7.001 7.001 0 011.05 8.84a.75.75 0 01.656-.834z"
})))), /*#__PURE__*/React.createElement("div", {
  className: "col-12 d-sm-flex d-xs-flex d-md-none"
}, props.result && /*#__PURE__*/React.createElement("div", {
  className: "primary-result-ready text-center"
}, /*#__PURE__*/React.createElement("span", {
  className: "agg-name"
}, props.result.primaryAggregate.name), /*#__PURE__*/React.createElement("span", {
  className: "agg-value"
}, props.result.primaryAggregate.value))));

const VoteResult = props => /*#__PURE__*/React.createElement("div", {
  className: "result-view mb-3"
}, /*#__PURE__*/React.createElement("div", {
  className: "clearfix"
}, /*#__PURE__*/React.createElement("div", {
  className: "float-right",
  style: {
    lineHeight: 1
  }
}, props.unmask && /*#__PURE__*/React.createElement("div", {
  id: "UnmaskBtn",
  onClick: props.onUnmask,
  className: "badge badge-danger"
}, "Unmask"), props.result.scores.map((score, i) => /*#__PURE__*/React.createElement("span", {
  key: i,
  className: "badge badge-score"
}, score)), props.result.range.overLimit && /*#__PURE__*/React.createElement("span", {
  className: "badge overlimit"
}, props.result.range.reason)), /*#__PURE__*/React.createElement("div", {
  className: "secondary-aggr badge float-left"
}, props.result.secondaryAggregate.name, " ", props.result.secondaryAggregate.value)));

const Progress = props => {
  const ratio = props.ratio + '%';
  var voted = props.voted - props.skipped;
  var skipmsg = voted + " voted";

  if (props.skipped > 0) {
    if (props.skipped === props.total) {
      skipmsg = "all skipped voting";
    } else {
      skipmsg = skipmsg + " / " + props.skipped + " skipped";
    }
  }

  return /*#__PURE__*/React.createElement("div", {
    className: "progress my-3"
  }, /*#__PURE__*/React.createElement("div", {
    className: "progress-bar",
    role: "progressbar",
    style: {
      width: ratio
    }
  }, /*#__PURE__*/React.createElement("span", {
    className: "skip-count"
  }, skipmsg)));
};

class FibBoard extends React.Component {
  constructor(props) {
    super(props);
    this.handleClick = this.handleClick.bind(this);
  }

  handleClick(e) {
    this.props.onVote(parseInt(e.target.dataset.score));
  }

  render() {
    const props = this.props;
    return /*#__PURE__*/React.createElement("div", {
      className: "fibboard"
    }, props.sequence.map((n, i) => /*#__PURE__*/React.createElement("button", {
      key: n + i,
      className: "btn btn-score",
      type: "button",
      onClick: this.handleClick,
      "data-score": n
    }, n)), /*#__PURE__*/React.createElement("button", {
      key: "unvote",
      className: "btn btn-score btn-score-control",
      type: "button",
      "data-score": "-1",
      onClick: this.handleClick
    }, "Unvote"), /*#__PURE__*/React.createElement("button", {
      key: "skip_voting",
      className: "btn btn-score btn-score-control",
      type: "button",
      "data-score": "-2",
      onClick: this.handleClick
    }, "Skip"));
  }

}

const UserScoreLine = props => /*#__PURE__*/React.createElement("div", {
  className: "user-vote",
  datauser: props.user
}, /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement("span", {
  className: "voter-name"
}, props.user), /*#__PURE__*/React.createElement("span", {
  className: "badge badge-score"
}, props.score), props.leader && /*#__PURE__*/React.createElement("div", {
  className: "leader-badge float-right"
}, "\u2605")), props.children);

class Voting extends React.Component {
  constructor(props) {
    super(props);
    this.delimiter = "|";
    this.fibonacci = new FibSeq({
      size: this.props.fibSize,
      bucketThreshold: this.props.fibBucketThreshold
    });
    this.handleVote = this.handleVote.bind(this);
    this.handleUnmask = this.handleUnmask.bind(this);
    this.handleClose = this.handleClose.bind(this);
  }

  handleVote(score) {
    api.sessionVote(score, () => toastr.success('Accepted', {
      timeOut: 100
    }));
  }

  handleUnmask() {
    api.sessionUnmask(() => toastr.success("Unmasked"));
  }

  handleClose() {
    api.sessionClose(() => toastr.success("Closed"));
  }

  decorateResult(result) {
    const copy = Object.assign({}, result);
    copy.scores = (result.scores || []).sort((a, b) => a - b);
    copy.range = this.fibonacci.isOutOfBucket(copy.scores);
    var aggregates = [['Fib:', this.fibonacci.closestFib(result.average)], ['Average:', result.average]];

    if (this.props.primaryAggregate == 'average') {
      aggregates = aggregates.reverse();
    }

    aggregates = aggregates.map(a => ({
      name: a[0],
      value: a[1]
    }));
    copy.primaryAggregate = aggregates[0];
    copy.secondaryAggregate = aggregates[1];
    return copy;
  }

  computeProgress(chain) {
    var progress = {
      voted: 0,
      total: 0
    };

    for (var v in chain.voters) {
      var score = chain.voters[v];
      if (score && score.length) progress.voted++;
      progress.total++;
    }

    progress.ratio = Math.round(100 * (progress.voted / (progress.total || 1)));
    progress.unvoted = progress.total - progress.voted;
    return progress;
  }

  render() {
    var chain = this.props.chain;
    var [name, color] = chain.name.split(this.delimiter);
    var progress = this.computeProgress(chain);
    var result = chain.result ? this.decorateResult(chain.result) : null;
    var unmask = !chain.unmasked && chain.leader == this.props.user.name;
    var voters = Object.keys(chain.voters).map(username => ({
      name: username,
      score: chain.voters[username],
      leader: chain.leader === username
    }));
    var voter = voters.find(v => v.name == this.props.user.name);
    var peers = voters.filter(p => p !== voter).sort();
    var masterLeader = voters.filter(v => v.leader).length == 0;
    return /*#__PURE__*/React.createElement("div", {
      id: "SessionState",
      className: color
    }, /*#__PURE__*/React.createElement(VotingHeader, {
      name: name,
      result: result,
      leader: masterLeader,
      onReset: this.props.onReset
    }), /*#__PURE__*/React.createElement(Progress, {
      ratio: progress.ratio,
      total: progress.total,
      voted: progress.voted,
      skipped: chain.skipped
    }), result && /*#__PURE__*/React.createElement(VoteResult, {
      result: result,
      unmask: unmask,
      onUnmask: this.handleUnmask
    }), /*#__PURE__*/React.createElement("div", null, voter && /*#__PURE__*/React.createElement(UserScoreLine, {
      user: voter.name,
      score: voter.score,
      leader: voter.leader
    }, /*#__PURE__*/React.createElement(FibBoard, {
      onVote: this.handleVote,
      sequence: this.fibonacci.getSequence()
    })), peers.map(p => /*#__PURE__*/React.createElement(UserScoreLine, {
      key: p.name,
      user: p.name,
      score: p.score,
      leader: p.leader
    }))), /*#__PURE__*/React.createElement("div", {
      className: "row mt-3"
    }, /*#__PURE__*/React.createElement("div", {
      className: "col-sm-4 offset-sm-4"
    }, /*#__PURE__*/React.createElement("button", {
      onClick: this.handleClose,
      className: "btn btn-sm btn-block btn-close"
    }, "Close"))));
  }

}

class SessionSubscriber {
  constructor(topic, token, handler) {
    this.url = 'ws://' + location.host + '/session/' + topic + '?authorization=' + (token || '');
    this.token = token;
    this.handler = handler;
    this.firstTime = true;
    this.maxRefereshWaitMs = 6000;
  }

  connect() {
    if (this.socket) {
      return;
    }

    var url = this.url;

    if (this.firstTime) {
      url += '&sync';
      this.firstTime = false;
    }

    this.socket = new WebSocket(url);
    this.socket.addEventListener('error', console.error);
    this.socket.addEventListener('close', () => {
      this.firstTime = true;
      this.socket = null;
      var randomMs = 5000 + parseInt(Math.random() * 3000);
      var waitMs = Math.min(this.maxRefereshWaitMs, randomMs);
      setTimeout(this.connect.bind(this), waitMs);
    });
    this.socket.addEventListener('message', event => {
      if (!event.data || !event.data.length) return;
      ClickSound.play();
      this.handler(JSON.parse(event.data));
    });
  }

}

;

class Session extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      session: null,
      showResetModal: false,
      loader: this.getLoadingMarker("Loading session")
    };
    this.stream = new SessionSubscriber('changes', this.props.user.token, this.handleIncomingUpdate.bind(this));
    this.handleOpen = this.handleOpen.bind(this);
    this.handleReset = this.handleReset.bind(this);
    this.handleCancelReset = this.handleCancelReset.bind(this);
    this.handleConfirmReset = this.handleConfirmReset.bind(this);
  }

  handleReset() {
    this.setState({
      showResetModal: true
    });
  }

  handleCancelReset() {
    this.setState({
      showResetModal: false
    });
  }

  handleConfirmReset() {
    api.sessionReset(() => toastr.success("Reseted"));
  }

  handleIncomingUpdate(session) {
    this.setState({
      session,
      showResetModal: false
    });
  }

  getLoadingMarker(message) {
    return {
      message
    };
  }

  handleOpen(voters) {
    this.setState({
      loader: this.getLoadingMarker("Openning session")
    });
    api.sessionOpen(voters.map(v => v.name), session => {
      this.setState({
        session,
        loader: null
      });
    }, this.fail.bind(this));
  }

  componentDidMount() {
    api.session(session => {
      this.setState({
        session,
        loader: null
      });
      this.stream.connect();
    }, this.fail.bind(this));
  }

  fail(res) {
    this.setState({
      error: res.error.error,
      loader: null
    });
  }

  render() {
    let component;

    if (this.state.loader) {
      component = /*#__PURE__*/React.createElement(Spinner, {
        message: this.state.loader.message
      });
    } else if (this.state.error) {
      component = /*#__PURE__*/React.createElement("div", {
        className: "alert alert-danger"
      }, "Encountered error: ", this.state.error, ". ", /*#__PURE__*/React.createElement("a", {
        href: "/"
      }, "Try to again"));
    } else if (this.state.session.chain) {
      component = /*#__PURE__*/React.createElement(Voting, {
        chain: this.state.session.chain,
        primaryAggregate: SessionOpts.primaryAggregate,
        fibSize: SessionOpts.fibonacci.size,
        fibBucketThreshold: SessionOpts.fibonacci.bucketThreshold,
        user: this.props.user,
        onReset: this.handleReset
      });
    } else {
      component = /*#__PURE__*/React.createElement(SessionStart, {
        onOpen: this.handleOpen
      });
    }

    return /*#__PURE__*/React.createElement("div", {
      className: "box view-max p-3 p-md-3 m-md-3 rounded"
    }, component, /*#__PURE__*/React.createElement(ReactBootstrap.Modal, {
      show: this.state.showResetModal,
      backdrop: "static",
      keyboard: false,
      size: "sm"
    }, /*#__PURE__*/React.createElement(ReactBootstrap.Modal.Body, null, /*#__PURE__*/React.createElement("p", {
      className: "text-center text-black"
    }, "Reset session?"), /*#__PURE__*/React.createElement("div", {
      className: "row"
    }, /*#__PURE__*/React.createElement("div", {
      className: "col-6"
    }, /*#__PURE__*/React.createElement("button", {
      type: "button",
      className: "btn btn-secondary float-left",
      onClick: this.handleCancelReset
    }, "Nope")), /*#__PURE__*/React.createElement("div", {
      className: "col-6 text-right"
    }, /*#__PURE__*/React.createElement("button", {
      type: "button",
      className: "btn btn-lemon",
      onClick: this.handleConfirmReset
    }, "Confirm!"))))));
  }

}

class Links extends React.Component {
  constructor(props) {
    super(props);
    this.state = this.makeLoadingState();
    this.handleRetry = this.handleRetry.bind(this);
  }

  makeStatus(message, retriable) {
    return {
      message,
      retriable
    };
  }

  handleRetry() {
    this.fetchLinks();
  }

  makeLoadingState() {
    return {
      links: null,
      status: this.makeStatus("Loading links...")
    };
  }

  fetchLinks() {
    this.setState(this.makeLoadingState());
    api.linkList(links => {
      this.setState({
        links,
        status: null
      });
    }, () => {
      this.setState({
        links: [],
        status: this.makeStatus("Failed to fetch links.", true)
      });
    });
  }

  componentDidMount() {
    this.fetchLinks();
  }

  render() {
    if (this.state.status) {
      return /*#__PURE__*/React.createElement("div", {
        className: "mt-3 px-4 px-lg-0 text-danger small"
      }, /*#__PURE__*/React.createElement("div", null, this.state.status.message), this.state.status.retriable && /*#__PURE__*/React.createElement("a", {
        href: "#",
        class: "text-decoration-none",
        onClick: this.handleRetry
      }, "Try again"));
    }

    const links = this.state.links.map((l, i) => {
      return /*#__PURE__*/React.createElement("div", {
        className: "mb-1",
        key: l.uri + i
      }, /*#__PURE__*/React.createElement("a", {
        href: l.uri,
        target: "_blank",
        title: l.uri,
        className: "text-truncate",
        style: {
          maxWidth: 350
        }
      }, l.display_name));
    });

    if (!links.length) {
      return /*#__PURE__*/React.createElement("span", null);
    }

    return /*#__PURE__*/React.createElement("div", {
      className: "p-3 small"
    }, /*#__PURE__*/React.createElement("h6", {
      className: "mb-2 text-muted"
    }, "Links"), links);
  }

}

ReactDOM.render( /*#__PURE__*/React.createElement(Session, {
  user: user
}), document.getElementById('SessionView'));
ReactDOM.render( /*#__PURE__*/React.createElement(Links, null), document.getElementById('Links'));
