// Copyright 2017 DENSSWeb Authors. All rights reserved.
//
// This file is part of DENSSWeb.
//
// DENSSWeb is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// DENSSWeb is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with DENSSWeb.  If not, see <http://www.gnu.org/licenses/>.
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : new P(function (resolve) { resolve(result.value); }).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
var __generator = (this && this.__generator) || function (thisArg, body) {
    var _ = { label: 0, sent: function() { if (t[0] & 1) throw t[1]; return t[1]; }, trys: [], ops: [] }, f, y, t;
    return { next: verb(0), "throw": verb(1), "return": verb(2) };
    function verb(n) { return function (v) { return step([n, v]); }; }
    function step(op) {
        if (f) throw new TypeError("Generator is already executing.");
        while (_) try {
            if (f = 1, y && (t = y[op[0] & 2 ? "return" : op[0] ? "throw" : "next"]) && !(t = t.call(y, op[1])).done) return t;
            if (y = 0, t) op = [0, t.value];
            switch (op[0]) {
                case 0: case 1: t = op; break;
                case 4: _.label++; return { value: op[1], done: false };
                case 5: _.label++; y = op[1]; op = [0]; continue;
                case 7: op = _.ops.pop(); _.trys.pop(); continue;
                default:
                    if (!(t = _.trys, t = t.length > 0 && t[t.length - 1]) && (op[0] === 6 || op[0] === 2)) { _ = 0; continue; }
                    if (op[0] === 3 && (!t || (op[1] > t[0] && op[1] < t[3]))) { _.label = op[1]; break; }
                    if (op[0] === 6 && _.label < t[1]) { _.label = t[1]; t = op; break; }
                    if (t && _.label < t[2]) { _.label = t[2]; _.ops.push(op); break; }
                    if (t[2]) _.ops.pop();
                    _.trys.pop(); continue;
            }
            op = body.call(thisArg, _);
        } catch (e) { op = [6, e]; y = 0; } finally { f = t = 0; }
        if (op[0] & 5) throw op[1]; return { value: op[0] ? op[1] : void 0, done: true };
    }
};
var LiteMol;
(function (LiteMol) {
    var Denss;
    (function (Denss) {
        var Plugin = LiteMol.Plugin;
        var Bootstrap = LiteMol.Bootstrap;
        var Transformer = Bootstrap.Entity.Transformer;
        // Download denss output as CCP4 and load the surface maps
        function loadMap(plugin, id, ccp4url) {
            return __awaiter(this, void 0, void 0, function () {
                var action, groupRef, group, denss, i, data, maxSigma, colors, scale, alpha, i, surface, sstyle;
                return __generator(this, function (_a) {
                    switch (_a.label) {
                        case 0:
                            action = plugin.createTransform();
                            groupRef = Bootstrap.Utils.generateUUID();
                            group = action.add(plugin.context.tree.root, Transformer.Basic.CreateGroup, { label: id, description: 'DENSS' }, { ref: groupRef });
                            denss = group
                                .then(Transformer.Data.Download, { url: ccp4url, type: 'Binary' })
                                .then(Transformer.Density.ParseData, { format: LiteMol.Core.Formats.Density.SupportedFormats.CCP4, id: id }, { isBinding: true, ref: 'denss-data' });
                            // Create 4 surface visuals with default styles
                            for (i = 1; i < 5; i++) {
                                denss.then(Transformer.Density.CreateVisual, { style: Bootstrap.Visualization.Density.Default.Style }, { ref: 'denss-s' + i });
                            }
                            // Render density map and surfaces and wait for transform to finish
                            return [4 /*yield*/, plugin.applyTransform(action)];
                        case 1:
                            // Render density map and surfaces and wait for transform to finish
                            _a.sent();
                            data = plugin.context.select('denss-data')[0];
                            maxSigma = Bootstrap.Utils.round(data.props.data.valuesInfo.max, 3);
                            colors = [0x0000FF, 0x008000, 0xFFFF00, 0xFF0000];
                            scale = [0.05, 0.10, 0.25, 0.50];
                            alpha = [0.10, 0.25, 0.25, 0.50];
                            i = 1;
                            _a.label = 2;
                        case 2:
                            if (!(i < 5)) return [3 /*break*/, 5];
                            surface = plugin.context.select('denss-s' + i)[0];
                            sstyle = Bootstrap.Visualization.Density.Style.create({
                                isoValue: scale[i - 1] * maxSigma,
                                isoValueType: Bootstrap.Visualization.Density.IsoValueType.Sigma,
                                color: LiteMol.Visualization.Color.fromHex(colors[i - 1]),
                                isWireframe: false,
                                transparency: { alpha: alpha[i - 1] }
                            });
                            return [4 /*yield*/, Transformer.Density.CreateVisual.create({ style: sstyle }, { ref: surface.ref }).update(plugin.context, surface).run()];
                        case 3:
                            _a.sent();
                            _a.label = 4;
                        case 4:
                            i++;
                            return [3 /*break*/, 2];
                        case 5:
                            // After all surfaces have been updated reset scene view (this resets the zoom)
                            plugin.command(Bootstrap.Command.Visual.ResetScene);
                            return [2 /*return*/];
                    }
                });
            });
        }
        // Grab job ID from HTML element
        var id = ((document.getElementById('jobid').value) || '').trim();
        // Grab CCP4 data URL from HTML element
        var ccp4url = ((document.getElementById('ccp4url').value) || '').trim();
        // Create LiteMol plugin
        var plugin = Plugin.create({
            target: '#app',
            viewportBackground: '#191919',
            layoutState: {
                hideControls: true,
                isExpanded: false
            },
            // XXX Enable this in production
            allowAnalytics: false
        });
        loadMap(plugin, id, ccp4url);
    })(Denss = LiteMol.Denss || (LiteMol.Denss = {}));
})(LiteMol || (LiteMol = {}));
