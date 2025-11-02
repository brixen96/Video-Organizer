You are an expert in the Go programming language.

# 📺 Video Storage AI – Project Specification

## My PC Specs & Important software
-	CPU: Intel Core i7 14700K
-	Memory: 32GB DDR5
-	Storage: App running on a M.2 NVME SAMSUNG 990 PRO 2TB
-	Ethernet: 1gbit
-	GPU: NVIDIA RTX 4080 Super
-	Latest NVIDIA Game Ready Drivers
-	ffmpeg version 8.0-full_build-www.gyan.dev
-	VLC Media Player 3.0.21 Vetinari
-	K-Lite Codec Pack 19.1.5 Full

## 📝 Overview
**Video Storage AI** is a Python-based Windows desktop application designed to act as a **video browser, video player, and AI-powered library organizer**.  
Its primary purpose is to provide a **YouTube-like browsing experience** for user-defined video folders while including powerful organization tools and an AI assistant for maintaining and improving large video libraries.

## 🎬 Context / Backstory
The user currently runs a Jellyfin server with multiple libraries:
- 🎥 **Movies & TV Shows**
- ✂️ **Edited Videos** (finished edits)
- 🗂️ **Backup Videos** (raw/original source files)

**The challenge:**  
Managing **500+ videos** across multiple folders where new content is constantly added.  
The current workflow using Windows File Explorer is functional but **slow, not scalable, and lacks intelligent organization**.

## 🎯 Goals
- Replace basic file browsing with a **video-focused interface**.  
- Integrate **metadata display, previews, and playback** directly in the app.  
- Provide **AI-powered assistance** for:  
- Suggesting naming conventions & folder structures.  
- Helping manage large amounts of new videos.  
- Planning future library improvements (e.g., tagging, categorizing, restructuring).  
- Enable **file and folder management** (rename, move, delete, organize) directly in the app.  
- Maintain an experience similar to **YouTube browsing**, but fully **local and personal**. 
- Turbo fast local caching system

## ⚙️ Core Features

### 📂 Video Browser UI
- YouTube-like interface for browsing videos.  
- Support for folders + subfolders.  
- Search and filter options.
- Whole App is Dark Themed.

### ▶️ Video Player
- Built-in video playback.  
- Display metadata (filename, size, codec, duration, resolution, etc.).  
- Thumbnails / previews for fast navigation.  

### 🤖 AI Companion
- Acts as an organizing assistant.  
- Suggests renaming conventions for files/folders.  
- Recommends folder structures for better scalability.  
- Helps identify duplicates, unfinished edits, or misplaced files.  
- Provides planning suggestions for library improvements.
- Live communication with the user

### 🛠️ Organization Tools
- Rename/move/delete videos without leaving the app.  
- Batch operations for newly downloaded files.  
- Drag-and-drop reorganization.  
- AI-guided bulk sorting.  

### 🚀 Scalability & Performance
- Must handle libraries with **500+ videos** efficiently.  
- Optimized folder scanning and metadata extraction.  

## 💻 Target Platform
- Python-based desktop app.  
- Runs on **Windows (priority)**.  
- Future potential for **cross-platform support**.  

## 🌟 Stretch Features (Optional for Later)
- 🏷️ Tagging system for custom video categories.  
- 🔗 Integration with existing Jellyfin libraries.  
- 🤖 Advanced AI features like automatic categorization or duplicate detection.  
- 🖼️ Preview generation (thumbnails, short clips).