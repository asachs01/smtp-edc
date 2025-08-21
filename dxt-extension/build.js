#!/usr/bin/env node

import fs from 'fs/promises';
import path from 'path';
import { exec } from 'child_process';
import { promisify } from 'util';
import { createWriteStream, createReadStream } from 'fs';
import archiver from 'archiver';
import { fileURLToPath } from 'url';

const execAsync = promisify(exec);
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

/**
 * Build and package SMTP-EDC Desktop Extension
 */
class DXTBuilder {
  constructor() {
    this.rootDir = __dirname;
    this.serverDir = path.join(this.rootDir, 'server');
    this.distDir = path.join(this.rootDir, 'dist');
    this.outputFile = path.join(this.distDir, 'smtp-edc.dxt');
  }

  async clean() {
    console.log('🧹 Cleaning previous build...');
    try {
      await fs.rm(this.distDir, { recursive: true, force: true });
    } catch (error) {
      // Directory might not exist
    }
    await fs.mkdir(this.distDir, { recursive: true });
  }

  async installDependencies() {
    console.log('📦 Installing dependencies...');
    const { stdout, stderr } = await execAsync('npm ci --production', {
      cwd: this.serverDir
    });
    
    if (stderr && !stderr.includes('npm WARN')) {
      console.error('Warning during npm install:', stderr);
    }
    console.log('✅ Dependencies installed');
  }

  async copyFiles() {
    console.log('📁 Copying files...');
    
    // Create temporary build directory
    const buildDir = path.join(this.distDir, 'build');
    await fs.mkdir(buildDir, { recursive: true });
    
    // Copy manifest
    await fs.copyFile(
      path.join(this.rootDir, 'manifest.json'),
      path.join(buildDir, 'manifest.json')
    );
    
    // Copy server directory with node_modules
    const serverBuildDir = path.join(buildDir, 'server');
    await this.copyDirectory(this.serverDir, serverBuildDir);
    
    // Copy icon if it exists
    try {
      await fs.copyFile(
        path.join(this.rootDir, 'icon.png'),
        path.join(buildDir, 'icon.png')
      );
    } catch (error) {
      console.log('ℹ️  No icon.png found, skipping');
    }
    
    console.log('✅ Files copied');
    return buildDir;
  }

  async copyDirectory(src, dest) {
    await fs.mkdir(dest, { recursive: true });
    const entries = await fs.readdir(src, { withFileTypes: true });
    
    for (const entry of entries) {
      const srcPath = path.join(src, entry.name);
      const destPath = path.join(dest, entry.name);
      
      if (entry.isDirectory()) {
        await this.copyDirectory(srcPath, destPath);
      } else {
        await fs.copyFile(srcPath, destPath);
      }
    }
  }

  async createArchive(buildDir) {
    console.log('🗜️  Creating DXT archive...');
    
    return new Promise((resolve, reject) => {
      const output = createWriteStream(this.outputFile);
      const archive = archiver('zip', {
        zlib: { level: 9 } // Maximum compression
      });

      output.on('close', () => {
        const size = (archive.pointer() / 1024).toFixed(2);
        console.log(`✅ DXT created: ${this.outputFile} (${size} KB)`);
        resolve();
      });

      archive.on('error', (err) => {
        reject(err);
      });

      archive.pipe(output);
      archive.directory(buildDir, false);
      archive.finalize();
    });
  }

  async cleanupBuildDir() {
    console.log('🧹 Cleaning up build directory...');
    const buildDir = path.join(this.distDir, 'build');
    await fs.rm(buildDir, { recursive: true, force: true });
  }

  async validateManifest() {
    console.log('🔍 Validating manifest...');
    const manifestPath = path.join(this.rootDir, 'manifest.json');
    const manifestContent = await fs.readFile(manifestPath, 'utf8');
    const manifest = JSON.parse(manifestContent);
    
    // Required fields
    const required = ['dxt_version', 'name', 'version', 'description', 'author', 'server'];
    for (const field of required) {
      if (!manifest[field]) {
        throw new Error(`Missing required field in manifest: ${field}`);
      }
    }
    
    // Validate server configuration
    if (!manifest.server.type || !manifest.server.entry_point) {
      throw new Error('Invalid server configuration in manifest');
    }
    
    console.log('✅ Manifest valid');
  }

  async generateChecksum() {
    console.log('🔐 Generating checksum...');
    const crypto = await import('crypto');
    const fileBuffer = await fs.readFile(this.outputFile);
    const hash = crypto.createHash('sha256');
    hash.update(fileBuffer);
    const checksum = hash.digest('hex');
    
    const checksumFile = this.outputFile.replace('.dxt', '.sha256');
    await fs.writeFile(checksumFile, `${checksum}  smtp-edc.dxt\n`);
    
    console.log(`✅ Checksum: ${checksum}`);
  }

  async build() {
    try {
      console.log('🚀 Building SMTP-EDC Desktop Extension...\n');
      
      await this.validateManifest();
      await this.clean();
      await this.installDependencies();
      const buildDir = await this.copyFiles();
      await this.createArchive(buildDir);
      await this.cleanupBuildDir();
      await this.generateChecksum();
      
      console.log('\n✨ Build complete!');
      console.log(`📦 Extension package: ${this.outputFile}`);
      console.log('\nTo install:');
      console.log('1. Open your AI assistant that supports DXT');
      console.log('2. Navigate to extensions/integrations settings');
      console.log('3. Click "Install from file" and select smtp-edc.dxt');
    } catch (error) {
      console.error('\n❌ Build failed:', error.message);
      process.exit(1);
    }
  }
}

// Run the build
const builder = new DXTBuilder();
builder.build();