const sigma = require('sigma');
const electron = require('electron').remote;
global.sigma = sigma;
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
const blue = '#0775FF';
const N = Object.keys(nw.Cells).length;
let layerInput = 0;
let layerMiddle = 0;
let layerOutput = 0;
let fanout = 10;
const isInput = tag => tag && tag.substring(0, 3) === 'in-';
const isMiddle = tag => !tag;

Object.keys(nw.Cells).forEach((cellId, i) => {
  const cell = nw.Cells[cellId];

  let x;
  let y;

  if (isInput(cell.Tag)) {
    layerInput += fanout;
    x = layerInput;
    y = 200;
  } else if (isMiddle(cell.Tag)) {
    layerMiddle += fanout;
    x = layerMiddle;
    y = 100;
  } else {
    layerOutput += fanout;
    x = layerOutput;
    y = 0;
  }
  fanout = -fanout;

    // x = 100 * Math.cos(2 * i * Math.PI / N); // random location
    // y = 100 * Math.sin(2 * i * Math.PI / N); // random location
            // : '#' + (Math.floor(Math.random() * 16777215).toString(16) + '000000').substr(0, 6),

  g.nodes.push({
    id: cell.ID,
    label: cell.Tag || cell.ID,
    size: Object.keys(cell.AxonSynapses).length + Object.keys(cell.DendriteSynapses).length,
    color: cell.Tag ?
            isInput(cell.Tag) ? blue : green
            : '#cccccc',
    x,
    y,
  });
});
Object.keys(nw.Synapses).forEach((synapseId) => {
  const synapse = nw.Synapses[synapseId];
  g.edges.push({
    id: synapse.ID,
    source: synapse.FromNeuronAxon,
    target: synapse.ToNeuronDendrite,
    type: 'curvedArrow',
  });
});

const s = new sigma({
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


window.start = () => {
  s.startForceAtlas2({
    worker: false,
        // linLogMode: false,
    outboundAttractionDistribution: true,
    adjustSizes: true,
        // edgeWeightInfluence: 0,
        // scalingRatio: 1,
        // strongGravityMode: false,
        // gravity: 1,
        // slowDown: 1,
        // barnesHutOptimize: false,
        // barnesHutTheta: 0.5,
        // startingIterations: 1,
        // iterationsPerRender: 1
  });
};

window.start();

window.stop = () => {
  s.stopForceAtlas2();
};
