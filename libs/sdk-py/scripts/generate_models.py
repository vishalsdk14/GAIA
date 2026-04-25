# Copyright 2026 GAIA Contributors
#
# Licensed under the MIT License.
# See the License for the specific language governing permissions and
# limitations under the License.

"""
This script automates the generation of Pydantic models from the GAIA JSON Schemas.
It ensures that the Python SDK maintains strict compatibility with the Go Kernel's
canonical data model without requiring manual maintenance of model files.
"""

import os
import subprocess
from pathlib import Path

def generate():
    schemas_dir = Path(__file__).parent.parent.parent.parent / "docs" / "specs" / "schemas"
    output_dir = Path(__file__).parent.parent / "gaia" / "models"
    
    print(f"Generating models from {schemas_dir}...")
    
    # We'll use datamodel-codegen to generate models from the directory
    # --input-file-type jsonschema
    # --output models.py
    
    # Header for the file
    header = "# Copyright 2026 GAIA Contributors\n# Auto-generated from JSON Schemas. DO NOT EDIT.\n\n"
    
    # Clear or create the file
    subprocess.run([
        "datamodel-codegen",
        "--input", str(schemas_dir),
        "--input-file-type", "jsonschema",
        "--output", str(output_dir),
        "--use-schema-description",
        "--field-constraints",
        "--use-standard-collections",
        "--base-class", "pydantic.BaseModel",
        "--allow-extra-fields"
    ], check=True)

    print(f"✅ Generated {output_dir}")

if __name__ == "__main__":
    generate()
