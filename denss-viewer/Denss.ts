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

namespace LiteMol.Denss {

    import Plugin = LiteMol.Plugin;
    import Bootstrap = LiteMol.Bootstrap;            
    import Transformer = Bootstrap.Entity.Transformer;

    // Download denss output as CCP4 and load the surface maps
    async function loadMap(plugin: Plugin.Controller, id: string, ccp4url: string) {
        let action = plugin.createTransform();

        let groupRef = Bootstrap.Utils.generateUUID();       
        let group = action.add(plugin.context.tree.root, Transformer.Basic.CreateGroup, { label: id, description: 'DENSS' }, { ref: groupRef })

        // Download binary density data in CCP4 format
        let denss = group
            .then(Transformer.Data.Download, { url: ccp4url, type: 'Binary' })
            .then(Transformer.Density.ParseData, { format: LiteMol.Core.Formats.Density.SupportedFormats.CCP4, id: id}, { isBinding: true, ref: 'denss-data' })

        // Create 4 surface visuals with default styles
        for (let i = 1; i < 5; i++) {
            denss.then(Transformer.Density.CreateVisual, { style: Bootstrap.Visualization.Density.Default.Style }, { ref: 'denss-s'+i });
        }

        // Render density map and surfaces and wait for transform to finish
        await plugin.applyTransform(action);

        // After the data is parsed, find max sigma and update the surfaces
        // with colors, sigma, and alpha values

        // Fetch parsed CCP4 data
        let data = plugin.context.select('denss-data')[0] as Bootstrap.Entity.Density.Data;

        // Find max sigma value
        let maxSigma = Bootstrap.Utils.round(data.props.data.valuesInfo.max, 3)

        let colors = [0x0000FF, 0x008000, 0xFFFF00, 0xFF0000] 
        let scale = [0.05, 0.10, 0.25, 0.50] 
        let alpha = [0.10, 0.25, 0.25, 0.50] 

        // Update surface styles
        for (let i = 1; i < 5; i++) {
            let surface = plugin.context.select('denss-s'+i)[0] as Bootstrap.Entity.Density.Visual;
            let sstyle = Bootstrap.Visualization.Density.Style.create({
                    isoValue: scale[i-1]*maxSigma,
                    isoValueType: Bootstrap.Visualization.Density.IsoValueType.Sigma,
                    color: LiteMol.Visualization.Color.fromHex(colors[i-1]), 
                    isWireframe: false,
                    transparency: { alpha: alpha[i-1]}
            });
            await Transformer.Density.CreateVisual.create({ style: sstyle }, { ref: surface.ref }).update(plugin.context, surface).run();
        }

        // After all surfaces have been updated reset scene view (this resets the zoom)
        plugin.command(Bootstrap.Command.Visual.ResetScene);
    }
        
    // Grab job ID from HTML element
    let id = (((document.getElementById('jobid') as HTMLInputElement).value) || '').trim();
    
    // Grab CCP4 data URL from HTML element
    let ccp4url = (((document.getElementById('ccp4url') as HTMLInputElement).value) || '').trim();

    // Create LiteMol plugin
    let plugin = Plugin.create({
        target: '#app',
        layoutState: {
            hideControls: true,
            isExpanded: false
        },

        // XXX Enable this in production
        allowAnalytics: false  
    });

    loadMap(plugin, id, ccp4url)
}
