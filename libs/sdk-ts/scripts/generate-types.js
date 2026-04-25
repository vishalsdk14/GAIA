/**
 * Copyright 2026 GAIA Contributors
 *
 * Licensed under the MIT License.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/**
 * This script automates the generation of TypeScript interfaces from the GAIA
 * JSON Schemas using quicktype. It ensures that the TypeScript SDK remains
 * perfectly synchronized with the Kernel's canonical data model.
 */

const {
    quicktype,
    InputData,
    JSONSchemaInput,
    JSONSchemaStore,
} = require("quicktype-core");
const fs = require('fs');
const path = require('path');

const schemasDir = path.join(__dirname, '../../../docs/specs/schemas');
const outputDir = path.join(__dirname, '../src');

async function quicktypeJSONSchema(targetLanguage, typeName, jsonSchemaString) {
    const schemaInput = new JSONSchemaInput(new JSONSchemaStore());
    await schemaInput.addSource({ name: typeName, schema: jsonSchemaString });

    const inputData = new InputData();
    inputData.addInput(schemaInput);

    return await quicktype({
        inputData,
        lang: targetLanguage,
        rendererOptions: {
            "just-types": "true",
        },
    });
}

async function generate() {
    const files = fs.readdirSync(schemasDir).filter(f => f.endsWith('.json'));
    let allTypes = '// Copyright 2026 GAIA Contributors\n// Auto-generated from JSON Schemas. DO NOT EDIT.\n\n';

    for (const file of files) {
        console.log(`Processing ${file}...`);
        const schemaContent = fs.readFileSync(path.join(schemasDir, file), 'utf8');
        const typeName = file.replace('.json', '').split('-').map(s => s.charAt(0).toUpperCase() + s.slice(1)).join('');
        
        const { lines } = await quicktypeJSONSchema("typescript", typeName, schemaContent);
        allTypes += lines.join("\n") + "\n";
    }

    fs.writeFileSync(path.join(outputDir, 'types.ts'), allTypes);
    console.log('✅ Generated src/types.ts');
}

generate().catch(console.error);
