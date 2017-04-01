const sigma = require('sigma');
const electron = require('electron').remote;
global.sigma = sigma;
require('./forceAtlas2/worker');
require('./forceAtlas2/supervisor');
require('sigma/plugins/sigma.plugins.animate/sigma.plugins.animate.js');
require('sigma/plugins/sigma.layout.noverlap/sigma.layout.noverlap.js');

const gunzip = require('zlib').gunzipSync
const fs = require('fs');

let NETWORK_FILE = process.env.NETWORK_FILE;

while (!NETWORK_FILE) {
    NETWORK_FILE = electron.dialog.showOpenDialog({
        properties: ['openFile'],
        filters: [{
                name: 'Compressed Network',
                extensions: ['nur']
            },
            {
                name: 'JSON Network',
                extensions: ['json']
            },
            {
                name: 'All Files',
                extensions: ['*']
            }
        ]
    })[0];
}

const rawText = gunzip(fs.readFileSync(NETWORK_FILE));
const nw = JSON.parse(rawText.toString('utf8'));
const g = {
    nodes: [],
    edges: []
};
const green = '#00FF2D';
const blue = '#0775FF';
const N = Object.keys(nw.Cells).length;
let layerInput = 0;
let layerMiddle = 0;
let layerOutput = 0;
const isInput = tag => tag && tag.substring(0, 3) === 'in-';
const isMiddle = tag => !tag;

Object.keys(nw.Cells).forEach((cellId, i) => {
    const cell = nw.Cells[cellId];

    let x;
    let y;

    if (isInput(cell.Tag)) {
        layerInput++;
        x = layerInput;
        y = 200;
    } else if (isMiddle(cell.Tag)) {
        layerMiddle++;
        x = layerMiddle;
        y = 100;
    } else {
        layerOutput++;
        x = layerOutput;
        y = 0;
    }

    // x = 100 * Math.cos(2 * i * Math.PI / N); // random location
    // y = 100 * Math.sin(2 * i * Math.PI / N); // random location

    g.nodes.push({
        id: cell.ID,
        label: cell.Tag || cell.ID,
        size: Object.keys(cell.AxonSynapses).length + Object.keys(cell.DendriteSynapses).length,
        color: cell.Tag ?
            isInput(cell.Tag) ? blue : green
            // : '#' + (Math.floor(Math.random() * 16777215).toString(16) + '000000').substr(0, 6),
            :
            '#cccccc',
        x,
        y
    });
});
Object.keys(nw.Synapses).forEach(synapseId => {
    const synapse = nw.Synapses[synapseId];
    g.edges.push({
        id: synapse.ID,
        source: synapse.FromNeuronAxon,
        target: synapse.ToNeuronDendrite
    });
});

// let i,
//     s,
//     o,
//     N = 1000,
//     E = 5000,
//     C = 5,
//     d = 0.5,
//     cs = [];
// // Generate the graph:
// for (i = 0; i < C; i++)
//     cs.push({
//         id: i,
//         nodes: [],
//         color: '#' + (
//             Math.floor(Math.random() * 16777215).toString(16) + '000000'
//         ).substr(0, 6)
//     });
// for (i = 0; i < N; i++) {
//     o = cs[(Math.random() * C) | 0];
//     g.nodes.push({
//         id: 'n' + i,
//         label: 'Node' + i,
//         x: 100 * Math.cos(2 * i * Math.PI / N),
//         y: 100 * Math.sin(2 * i * Math.PI / N),
//         size: Math.random(),
//         color: o.color
//     });
//     o.nodes.push('n' + i);
// }
// for (i = 0; i < E; i++) {
//     if (Math.random() < 1 - d)
//         g.edges.push({
//             id: 'e' + i,
//             source: 'n' + ((Math.random() * N) | 0),
//             target: 'n' + ((Math.random() * N) | 0)
//         });
//     else {
//         o = cs[(Math.random() * C) | 0]
//         g.edges.push({
//             id: 'e' + i,
//             source: o.nodes[(Math.random() * o.nodes.length) | 0],
//             target: o.nodes[(Math.random() * o.nodes.length) | 0]
//         });
//     }
// }

var s = new sigma({
    graph: g,
    container: 'graph-container',
    settings: {
        drawEdges: true,

        minNodeSize: 2,
        maxNodeSize: 10,
        minEdgeSize: 1,
        maxEdgeSize: 10,
    }
});

// Start the ForceAtlas2 algorithm:
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

// setTimeout(() => {
// s.stopForceAtlas2();
// }, 10000);
