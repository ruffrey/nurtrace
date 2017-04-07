const Sigma = require('sigma');
const electron = require('electron').remote; // eslint-disable-line

global.sigma = Sigma;
require('./forceAtlas2/worker');
require('./forceAtlas2/supervisor');
require('sigma/plugins/sigma.plugins.animate/sigma.plugins.animate.js');
require('sigma/plugins/sigma.layout.noverlap/sigma.layout.noverlap.js');
require('sigma/src/renderers/canvas/sigma.canvas.edges.curvedArrow.js');

const gunzip = require('zlib').gunzipSync;
const fs = require('fs');

let NETWORK_FILE = process.env.NETWORK_FILE;

while (!NETWORK_FILE) {
  NETWORK_FILE = electron.dialog.showOpenDialog({
    properties: ['openFile'],
    filters: [{
      name: 'Compressed Network',
      extensions: ['nur'],
    },
    {
      name: 'JSON Network',
      extensions: ['json'],
    },
    {
      name: 'All Files',
      extensions: ['*'],
    },
    ],
  })[0];
}

const rawText = gunzip(fs.readFileSync(NETWORK_FILE));
const nw = JSON.parse(rawText.toString('utf8'));
const g = {
  nodes: [],
  edges: [],
};
const green = '#00FF2D';
// const blue = '#0775FF';
const red = '#A62A2A';
const lightGrey = '#aaaaaa';
const darkGrey = '#444444';
const black = '#000000';
const white = '#ffffff';

const colorInhibitorySynapse = red;
const colorExcitatorysynapse = green;
const colorInputCell = white;
const colorMiddleCell = lightGrey;
const colorOutputCell = darkGrey;


let layerInput = 0;
let layerMiddle = 0;
let layerMiddleDepth = 1;
let layerOutput = 0;
let fanout = 10;
const isInput = tag => tag && tag.substring(0, 3) === 'in-';
const isMiddle = tag => !tag;

Object.keys(nw.Cells).forEach((cellId) => {
  const cell = nw.Cells[cellId];

  let x;
  let y;
  let color;

  if (isInput(cell.Tag)) {
    layerInput = -layerInput + fanout;
    x = layerInput;
    y = 1200;
  } else if (isMiddle(cell.Tag)) {
    layerMiddle = -layerMiddle + fanout;
    if (layerMiddle > 400) {
      layerMiddleDepth += (2 * fanout);
      layerMiddle = 0;
    }
    x = layerMiddle;
    y = layerMiddleDepth;
  } else {
    layerOutput = -layerOutput + fanout;
    x = layerOutput;
    y = -200;
  }
  fanout = -fanout;

    // x = 100 * Math.cos(2 * i * Math.PI / N); // random location
    // y = 100 * Math.sin(2 * i * Math.PI / N); // random location
            // : '#' + (Math.floor(Math.random() * 16777215).toString(16) + '000000').substr(0, 6),
  if (cell.Tag) {
    if (isInput(cell.Tag)) {
      color = colorInputCell;
    } else {
      color = colorOutputCell;
    }
  } else {
    color = colorMiddleCell;
  }

  g.nodes.push({
    id: cell.ID,
    label: cell.Tag || cell.ID,
    size: Object.keys(cell.AxonSynapses).length + Object.keys(cell.DendriteSynapses).length,
    color,
    x,
    y,
  });
});
Object.keys(nw.Synapses).forEach((synapseId) => {
  const synapse = nw.Synapses[synapseId];
  let color;
  if (synapse.Millivolts > 0) {
    color = colorExcitatorysynapse;
  } else if (synapse.Millivolts < 0) {
    color = colorInhibitorySynapse;
  } else {
    color = black;
  }

  g.edges.push({
    id: synapse.ID,
    source: synapse.FromNeuronAxon,
    target: synapse.ToNeuronDendrite,
    // type: 'curvedArrow',
    color,
  });
});

const s = new Sigma({
  graph: g,
  renderer: {
    container: 'graph-container',
    type: 'canvas',
  },
  settings: {
    drawEdges: true,

    minNodeSize: 2,
    maxNodeSize: 10,
    minEdgeSize: 1,
    maxEdgeSize: 10,
  },
});
window.s = s;

let lastSelectedPathNode;
const clearLastNode = () => {
  lastSelectedPathNode = null;
  document.getElementById('lastNodeDisplay').innerHTML = '';
};
const setLastNode = (cellId) => {
  lastSelectedPathNode = cellId;
  let tag = '';
  if (nw.Cells[cellId].Tag) {
    tag += ` (${nw.Cells[cellId].Tag})`;
  }
  document.getElementById('lastNodeDisplay').innerHTML = cellId + tag;
};

function selectPath(e) {
  console.log('select path', e);
  window.stop();

  const maxDepth = parseInt(document.getElementById('depth').value, 10);
  const alreadyWalked = {};
  const directionValue = document.getElementById('direction').value;
  const shouldSearchForward = directionValue === 'both' || directionValue === 'fwd';
  const shouldSearchBackward = directionValue === 'both' || directionValue === 'back';

  if (e) {
    setLastNode(e.data.node.id);
  }
  // Walk up and down just to see all connections
  // from this cell's perspective.
  const walk = (nodeId, depth) => {
    if (alreadyWalked[nodeId]) {
      return;
    }
    depth++;
    if (depth > maxDepth) {
      return;
    }
    const cell = nw.Cells[nodeId];
    const walkNext = [];
    alreadyWalked[nodeId] = true;
    if (shouldSearchForward) {
      Object.keys(cell.AxonSynapses).forEach((synapseId) => {
        const synapse = nw.Synapses[synapseId];
        // console.log(synapseId, synapse.ToNeuronDendrite);
        walkNext.push(synapse.ToNeuronDendrite);
      });
    }
    if (shouldSearchBackward) {
      Object.keys(cell.DendriteSynapses).forEach((synapseId) => {
        const synapse = nw.Synapses[synapseId];
                // console.log(synapseId, synapse.FromNeuronAxon);
        walkNext.push(synapse.FromNeuronAxon);
      });
    }

    walkNext.forEach(n => walk(n, depth));
  };
  walk(lastSelectedPathNode, 0);

  // only show nodes that were walked
  s.graph.nodes().forEach((n) => {
    if (alreadyWalked[n.id]) {
      n.hidden = false;
    } else {
      n.hidden = true;
    }
  });
  s.refresh();
  console.log('done walking path', lastSelectedPathNode, ',',
        Object.keys(alreadyWalked).length, '/', g.nodes.length, 'cells');
}
s.bind('clickNode', selectPath);

window.start = () => {
  s.startForceAtlas2({
    worker: false,
    linLogMode: true, // maybe?
    outboundAttractionDistribution: true,
    adjustSizes: true,
        // edgeWeightInfluence: 0,
        // scalingRatio: 1,
        // strongGravityMode: false,
    gravity: 10,
        // slowDown: 1,
        // barnesHutOptimize: false,
        // barnesHutTheta: 0.5,
        // startingIterations: 1,
    iterationsPerRender: 10,
  });
};

// window.start();

window.changeDepthIfSelectedPath = () => {
  if (!lastSelectedPathNode) {
    return;
  }
  selectPath();
};

window.stop = () => {
  s.stopForceAtlas2();
};

window.clearPathSelection = () => {
  clearLastNode();
  s.graph.nodes().forEach((n) => {
    n.hidden = false;
  });
  s.refresh();
};

window.hideAll = () => {
  s.graph.nodes().forEach((n) => {
    n.hidden = true;
  });
  s.refresh();
};
