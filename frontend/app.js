
const closeDetailsPanelAndPauseVideos = () => {
    const performerDetailsPanel = document.getElementById('performer-details-panel');
    const performerCarouselVideo = performerDetailsPanel ? performerDetailsPanel.querySelector('.performer-carousel video') : null;
    if (performerDetailsPanel) {
        performerDetailsPanel.classList.remove('active');
        if (performerCarouselVideo) {
            performerCarouselVideo.pause();
            performerCarouselVideo.src = '';
        }
    }

    const videoModal = document.getElementById('video-modal');
    const modalVideoPlayer = document.getElementById('modal-video-player');
    if (videoModal && modalVideoPlayer) {
        videoModal.style.display = 'none';
        modalVideoPlayer.pause();
        modalVideoPlayer.src = '';
    }
};

document.addEventListener('DOMContentLoaded', () => {
    const videoGrid = document.getElementById('video-grid');
    const modal = document.getElementById('video-modal');
    const modalVideoPlayer = document.getElementById('modal-video-player');
    const modalVideoDetails = document.getElementById('modal-video-details');
    const closeButton = document.querySelector('.close-button');
    const settingsPage = document.getElementById('settings-page');
    const scenesPage = document.getElementById('scenes-page');

    let currentVideo = null;
    let allVideos = []; // Global variable to store all video data

    const handlePageVisibility = (pageId) => {
        const pages = document.querySelectorAll('.page');
        pages.forEach(page => page.classList.remove('active'));

        const activePage = document.getElementById(pageId);
        if (activePage) {
            activePage.classList.add('active');
        }

        // Explicitly hide performer-details-panel if not on performers page
        if (pageId !== 'performers-page') {
            closeDetailsPanelAndPauseVideos();
        }

        if (pageId === 'scenes-page') {
            fetchAndRenderGrid();
        }
    };

    const handleSettingsTabs = () => {
        const tabButtons = settingsPage.querySelectorAll('.tab-button');
        const tabPanes = settingsPage.querySelectorAll('.tab-pane');

        tabButtons.forEach(button => {
            button.addEventListener('click', () => {
                tabButtons.forEach(btn => btn.classList.remove('active'));
                button.classList.add('active');

                tabPanes.forEach(pane => pane.classList.add('hidden'));
                const targetTab = document.getElementById(button.dataset.tab);
                if (targetTab) {
                    targetTab.classList.remove('hidden');
                }
            });
        });

        // Activate default tab
        const defaultTabButton = settingsPage.querySelector('.tab-button.active');
        if (defaultTabButton) {
            defaultTabButton.click();
        }
    };

    const router = () => {
        const hash = window.location.hash || '#scenes';
        const pageId = hash.substring(1) + '-page';
        handlePageVisibility(pageId);

        if (hash === '#settings') {
            handleSettingsTabs();
        } else if (hash === '#logs') {
            handleLogsTabs();
        } else if (hash === '#performers') {
            handlePerformersPage();
        }
    };

    const handlePerformersPage = () => {
        const performerWall = document.getElementById('performer-wall');
        const sortBySelect = document.getElementById('performer-sort-by');
        const filterZooSelect = document.getElementById('performer-filter-zoo');
        const ageMinInput = document.getElementById('age-min');
        const ageMaxInput = document.getElementById('age-max');
        const ageMinValueSpan = document.getElementById('age-min-value');
        const ageMaxValueSpan = document.getElementById('age-max-value');

        const heightMinInput = document.getElementById('height-min');
        const heightMaxInput = document.getElementById('height-max');
        const heightMinValueSpan = document.getElementById('height-min-value');
        const heightMaxValueSpan = document.getElementById('height-max-value');

        const cupMinSelect = document.getElementById('cup-min');
        const cupMaxSelect = document.getElementById('cup-max');

        const applyFiltersButton = document.getElementById('apply-filters-button');
        const resetFiltersButton = document.getElementById('reset-filters-button');

        let allPerformers = [];
        let minAgeGlobal = 0;
        let maxAgeGlobal = 0;
        let minHeightGlobal = 0;
        let maxHeightGlobal = 0;
        let allCupSizes = [];

        const closePerformerDetailsPanel = () => {
            closeDetailsPanelAndPauseVideos();
        };

        const cupSizeOrder = (cup) => {
            const order = ['AA', 'A', 'B', 'C', 'D', 'DD', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'];
            const upperCup = (cup || '').toUpperCase();
            let index = order.indexOf(upperCup);
            if (index !== -1) {
                return index;
            }
            if (upperCup.startsWith('DD')) return order.indexOf('DD') + (upperCup.length - 2) * 0.5;
            if (upperCup.startsWith('EE')) return order.indexOf('E') + (upperCup.length - 2) * 0.5;
            return -1;
        };

        const parseHeight = (heightStr) => {
            if (!heightStr) return 0;
            const match = heightStr.match(/(\d+)/);
            return match ? parseInt(match[1], 10) : 0;
        };

        const renderPerformers = (performers) => {
            performerWall.innerHTML = ''; // Clear previous content
            if (performers.length === 0) {
                performerWall.textContent = 'No performers found matching the criteria.';
                return;
            }
            performers.forEach(performer => {
                const currentPerformerName = performer.name; // Capture the name
                const performerItem = document.createElement('div');
                performerItem.className = 'performer-item';
                performerItem.innerHTML = `
                    <div class="performer-preview-container">
                        ${performer.default_preview 
                            ? `<video src="/performer-previews/${performer.default_preview}" loop muted class="performer-thumbnail"></video>` 
                            : (performer.previews && performer.previews.length > 0 
                                ? `<video src="/performer-previews/${performer.previews[0]}" loop muted class="performer-thumbnail"></video>` 
                                : `<img src="https://via.placeholder.com/150" alt="${performer.name}" class="performer-thumbnail">`)
                        }
                    </div>
                    <div class="performer-info">
                        <p class="performer-name">${performer.name}</p>
                        <p class="performer-scene-count">Scenes: ${performer.scene_count}</p>
                    </div>
                `;
                const videoElement = performerItem.querySelector('video');
                if (videoElement) {
                    performerItem.addEventListener('mouseenter', () => {
                        videoElement.play();
                    });
                    performerItem.addEventListener('mouseleave', () => {
                        videoElement.pause();
                        videoElement.currentTime = 0; // Reset video to start
                    });
                }
                performerItem.addEventListener('click', () => {
                    displayPerformerDetails(currentPerformerName);
                });
                performerWall.appendChild(performerItem);
            });
        };

        const applyFiltersAndSort = () => {
            let filteredPerformers = [...allPerformers];

            // Filters
            const minAge = parseInt(ageMinInput.value);
            const maxAge = parseInt(ageMaxInput.value);
            const minHeight = parseInt(heightMinInput.value);
            const maxHeight = parseInt(heightMaxInput.value);
            const minCupIndex = cupMinSelect.value === '' ? -1 : cupSizeOrder(cupMinSelect.value);
            const maxCupIndex = cupMaxSelect.value === '' ? Infinity : cupSizeOrder(cupMaxSelect.value);

            filteredPerformers = filteredPerformers.filter(p => {
                const performerAge = parseInt(p.age);
                const performerHeight = parseHeight(p.appearance?.height);
                const performerCupIndex = cupSizeOrder(p.appearance?.cup);

                // Age filter
                let ageMatch = true;
                // If the age filter is NOT set to its widest possible range (minAgeGlobal to maxAgeGlobal)
                if (!(minAge === minAgeGlobal && maxAge === maxAgeGlobal)) {
                    if (!isNaN(performerAge)) { // If performer has age
                        ageMatch = performerAge >= minAge && performerAge <= maxAge;
                    } else { // Performer has no age, so it doesn't match if a range is set
                        ageMatch = false;
                    }
                }

                // Height filter
                let heightMatch = true;
                // If the height filter is NOT set to its widest possible range (minHeightGlobal to maxHeightGlobal)
                if (!(minHeight === minHeightGlobal && maxHeight === maxHeightGlobal)) {
                    if (!isNaN(performerHeight) && performerHeight > 0) { // If performer has height
                        heightMatch = performerHeight >= minHeight && performerHeight <= maxHeight;
                    } else { // Performer has no height, so it doesn't match if a range is set
                        heightMatch = false;
                    }
                }

                // Cup filter
                let cupMatch = true;
                // If the cup filter is NOT set to 'All' for both min and max
                if (cupMinSelect.value !== '' || cupMaxSelect.value !== '') {
                    if (performerCupIndex !== -1) { // If performer has cup size
                        cupMatch = performerCupIndex >= minCupIndex && performerCupIndex <= maxCupIndex;
                    } else { // Performer has no cup size, so it doesn't match if a range is set
                        cupMatch = false;
                    }
                }

                // Zoo filter
                const zoo = filterZooSelect.value;
                let zooMatch = true;
                if (zoo === 'yes') {
                    zooMatch = p.zoo === 'true';
                } else if (zoo === 'no') {
                    zooMatch = p.zoo !== 'true';
                }

                return ageMatch && cupMatch && heightMatch && zooMatch;
            });

            // Sort
            const sortBy = sortBySelect.value;
            switch (sortBy) {
                case 'name-asc':
                    filteredPerformers.sort((a, b) => a.name.localeCompare(b.name));
                    break;
                case 'name-desc':
                    filteredPerformers.sort((a, b) => b.name.localeCompare(a.name));
                    break;
                case 'age-desc':
                    filteredPerformers.sort((a, b) => (parseInt(b.age) || 0) - (parseInt(a.age) || 0));
                    break;
                case 'age-asc':
                    filteredPerformers.sort((a, b) => (parseInt(a.age) || 0) - (parseInt(b.age) || 0));
                    break;
                case 'cup-desc':
                    filteredPerformers.sort((a, b) => cupSizeOrder(b.appearance?.cup) - cupSizeOrder(a.appearance?.cup));
                    break;
                case 'cup-asc':
                    filteredPerformers.sort((a, b) => cupSizeOrder(a.appearance?.cup) - cupSizeOrder(b.appearance?.cup));
                    break;
                case 'height-desc':
                    filteredPerformers.sort((a, b) => parseHeight(b.appearance?.height) - parseHeight(a.appearance?.height));
                    break;
                case 'height-asc':
                    filteredPerformers.sort((a, b) => parseHeight(a.appearance?.height) - parseHeight(a.appearance?.height));
                    break;
                case 'rating-desc':
                    filteredPerformers.sort((a, b) => (b.rating || 0) - (a.rating || 0));
                    break;
                case 'rating-asc':
                    filteredPerformers.sort((a, b) => (a.rating || 0) - (b.rating || 0));
                    break;
                case 'total-views-desc':
                    filteredPerformers.sort((a, b) => (b.total_views || 0) - (a.total_views || 0));
                    break;
                case 'total-views-asc':
                    filteredPerformers.sort((a, b) => (a.total_views || 0) - (b.total_views || 0));
                    break;
            }

            renderPerformers(filteredPerformers);
        };

        const resetFilters = () => {
            // Reset Age sliders
            ageMinInput.value = minAgeGlobal;
            ageMinValueSpan.textContent = minAgeGlobal;
            ageMaxInput.value = maxAgeGlobal;
            ageMaxValueSpan.textContent = maxAgeGlobal;

            // Reset Height sliders
            heightMinInput.value = minHeightGlobal;
            heightMinValueSpan.textContent = minHeightGlobal;
            heightMaxInput.value = maxHeightGlobal;
            heightMaxValueSpan.textContent = maxHeightGlobal;

            // Reset Cup size dropdowns
            cupMinSelect.value = '';
            cupMaxSelect.value = '';

            // Reset Zoo filter
            filterZooSelect.value = 'all';

            // Reset Sort By
            sortBySelect.value = 'name-asc';

            applyFiltersAndSort(); // Apply filters after reset
        };

        fetch('/api/performers')
            .then(response => response.json())
            .then(performers => {
                allPerformers = performers;

                // Initialize global min/max values for sliders
                const ages = performers.map(p => parseInt(p.age)).filter(age => !isNaN(age));
                minAgeGlobal = Math.min(...ages);
                maxAgeGlobal = Math.max(...ages);

                const heights = performers.map(p => parseHeight(p.appearance?.height)).filter(h => h > 0);
                minHeightGlobal = Math.min(...heights);
                maxHeightGlobal = Math.max(...heights);

                allCupSizes = [...new Set(performers.map(p => p.appearance?.cup).filter(cup => cup && cupSizeOrder(cup) !== -1))];
                allCupSizes.sort((a, b) => cupSizeOrder(a) - cupSizeOrder(b));

                // Set min/max for age and height sliders
                ageMinInput.min = minAgeGlobal;
                ageMinInput.max = maxAgeGlobal;
                ageMaxInput.min = minAgeGlobal;
                ageMaxInput.max = maxAgeGlobal;

                heightMinInput.min = minHeightGlobal;
                heightMinInput.max = maxHeightGlobal;
                heightMaxInput.min = minHeightGlobal;
                heightMaxInput.max = maxHeightGlobal;

                // Populate cup size dropdowns
                const populateCupDropdown = (selectElement) => {
                    selectElement.innerHTML = '<option value="">All</option>';
                    allCupSizes.forEach(cup => {
                        const option = document.createElement('option');
                        option.value = cup;
                        option.textContent = cup;
                        selectElement.appendChild(option);
                    });
                };

                populateCupDropdown(cupMinSelect);
                populateCupDropdown(cupMaxSelect);

                // Set initial slider values and text
                ageMinInput.value = minAgeGlobal;
                ageMinValueSpan.textContent = minAgeGlobal;
                ageMaxInput.value = maxAgeGlobal;
                ageMaxValueSpan.textContent = maxAgeGlobal;

                heightMinInput.value = minHeightGlobal;
                heightMinValueSpan.textContent = minHeightGlobal;
                heightMaxInput.value = maxHeightGlobal;
                heightMaxValueSpan.textContent = maxHeightGlobal;

                resetFilters(); // Initialize filters to default and display all performers
            })
            .catch(error => {
                console.error('Error fetching performers:', error);
                performerWall.textContent = 'Failed to load performers.';
            });

        // Event Listeners for filters and sort (no immediate apply)
        ageMinInput.addEventListener('input', () => {
            ageMinValueSpan.textContent = ageMinInput.value;
            if (parseInt(ageMinInput.value) > parseInt(ageMaxInput.value)) {
                ageMaxInput.value = ageMinInput.value;
                ageMaxValueSpan.textContent = ageMinInput.value;
            }
        });
        ageMaxInput.addEventListener('input', () => {
            ageMaxValueSpan.textContent = ageMaxInput.value;
            if (parseInt(ageMaxInput.value) < parseInt(ageMinInput.value)) {
                ageMinInput.value = ageMaxInput.value;
                ageMinValueSpan.textContent = ageMaxInput.value;
            }
        });
        heightMinInput.addEventListener('input', () => {
            heightMinValueSpan.textContent = heightMinInput.value;
            if (parseInt(heightMinInput.value) > parseInt(heightMaxInput.value)) {
                heightMaxInput.value = heightMinInput.value;
                heightMaxValueSpan.textContent = heightMinInput.value;
            }
        });
        heightMaxInput.addEventListener('input', () => {
            heightMaxValueSpan.textContent = heightMaxInput.value;
            if (parseInt(heightMaxInput.value) < parseInt(heightMinInput.value)) {
                heightMinInput.value = heightMaxInput.value;
                heightMinValueSpan.textContent = heightMaxInput.value;
            }
        });

        // Event listeners for buttons
        applyFiltersButton.addEventListener('click', applyFiltersAndSort);
        resetFiltersButton.addEventListener('click', resetFilters);

        // Initial render is handled by resetFilters() after data fetch
        sortBySelect.addEventListener('change', applyFiltersAndSort);
        filterZooSelect.addEventListener('change', applyFiltersAndSort);
    };

    const displayPerformerDetails = (performerName) => {
        const performerDetailsPanel = document.getElementById('performer-details-panel');
        const performerCarousel = document.getElementById('performer-carousel'); // Get carousel element
        const performerProfileContent = document.getElementById('performer-profile-content');
        const performerScenesContent = document.getElementById('performer-scenes-content');
        const performerAppearanceContent = document.getElementById('performer-appearance-content');
        const performerTagsContent = document.getElementById('performer-tags-content');
        const performerBiosContent = document.getElementById('performer-bios-content');
        const performerOtherInfoContent = document.getElementById('performer-other-info-content');
        const closeDetailsButton = performerDetailsPanel.querySelector('.close-details-button');

        performerDetailsPanel.classList.add('active');
        performerProfileContent.innerHTML = 'Loading profile...';
        performerScenesContent.innerHTML = 'Loading scenes...';
        performerAppearanceContent.innerHTML = 'Loading appearance...';
        performerTagsContent.innerHTML = 'Loading tags...';
        performerBiosContent.innerHTML = 'Loading bios...';
        performerOtherInfoContent.innerHTML = 'Loading other info...';
        performerCarousel.innerHTML = ''; // Clear carousel content

        // Close button for details panel
        closeDetailsButton.onclick = () => {
            closeDetailsPanelAndPauseVideos();
        };

        // Handle tabs within the details panel
        const tabButtons = performerDetailsPanel.querySelectorAll('.tab-buttons .tab-button');
        const tabPanes = performerDetailsPanel.querySelectorAll('.tab-content .tab-pane');

        tabButtons.forEach(button => {
            button.onclick = () => {
                tabButtons.forEach(btn => btn.classList.remove('active'));
                button.classList.add('active');

                tabPanes.forEach(pane => pane.classList.add('hidden'));
                const targetTab = document.getElementById(button.dataset.tab);
                if (targetTab) {
                    targetTab.classList.remove('hidden');
                }
            };
        });

        // Fetch performer details
        fetch(`/api/performers/${performerName}`)
            .then(response => response.json())
            .then(performer => {
                // --- Populate Carousel ---
                if (performer.previews && performer.previews.length > 0) {
                    const mainPreviewContainer = document.createElement('div');
                    mainPreviewContainer.className = 'main-preview-container';
                    performerCarousel.appendChild(mainPreviewContainer);

                    const thumbnailNav = document.createElement('div');
                    thumbnailNav.className = 'thumbnail-nav';
                    performerCarousel.appendChild(thumbnailNav);

                    // Function to render a preview in the main display
                    const renderMainPreview = (previewUrl) => {
                        mainPreviewContainer.innerHTML = ''; // Clear previous main preview
                        const video = document.createElement('video');
                        video.src = `/performer-previews/${previewUrl}`;
                        video.controls = true;
                        video.autoplay = true; // Autoplay the main focused video
                        video.loop = true;
                        video.muted = true;
                        mainPreviewContainer.appendChild(video);
                    };

                    // Render the first preview initially
                    renderMainPreview(performer.default_preview || performer.previews[0]);

                    // Populate thumbnail navigation
                    performer.previews.forEach(previewUrl => {
                        const thumbItem = document.createElement('div');
                        thumbItem.className = 'thumbnail-item';
                        // Add active class if it's the default preview
                        if (previewUrl === performer.default_preview) {
                            thumbItem.classList.add('active');
                        }

                        const video = document.createElement('video');
                        video.src = `/performer-previews/${previewUrl}`;
                        video.muted = true;
                        thumbItem.appendChild(video);
                        thumbItem.addEventListener('mouseenter', () => { video.play(); });
                        thumbItem.addEventListener('mouseleave', () => { video.pause(); video.currentTime = 0; });
                        
                        thumbItem.addEventListener('click', () => {
                            renderMainPreview(previewUrl);
                        });

                        // Context menu for setting default preview
                        thumbItem.addEventListener('contextmenu', (e) => {
                            e.preventDefault(); // Prevent default browser context menu
                            const contextMenu = document.getElementById('context-menu');
                            const setAsDefaultButton = document.getElementById('set-as-default-button');

                            contextMenu.style.left = `${e.pageX}px`;
                            contextMenu.style.top = `${e.pageY}px`;
                            contextMenu.style.display = 'block';

                            setAsDefaultButton.onclick = async () => {
                                try {
                                    const response = await fetch(`/api/performers/${performer.name}/set-default-preview`, {
                                        method: 'POST',
                                        headers: { 'Content-Type': 'application/json' },
                                        body: JSON.stringify({ previewUrl: previewUrl }),
                                    });
                                    if (!response.ok) {
                                        throw new Error(`HTTP error! status: ${response.status}`);
                                    }
                                    const result = await response.json();
                                    console.log(result.message);

                                    // Update UI to reflect new default
                                    performer.default_preview = previewUrl; // Update local performer object
                                    thumbnailNav.querySelectorAll('.thumbnail-item').forEach(item => item.classList.remove('active'));
                                    thumbItem.classList.add('active');

                                } catch (error) {
                                    console.error('Error setting default preview:', error);
                                    alert('Failed to set default preview.');
                                }
                                contextMenu.style.display = 'none';
                            };

                            const fetchMetadataButton = document.getElementById('fetch-metadata-button');
                            fetchMetadataButton.onclick = async () => {
                                try {
                                    const response = await fetch(`/api/performers/${performer.name}/fetch-metadata`, {
                                        method: 'POST',
                                    });
                                    if (!response.ok) {
                                        throw new Error(`HTTP error! status: ${response.status}`);
                                    }
                                    const result = await response.json();
                                    console.log(result.message);
                                    alert(result.message);
                                    // Refresh performer details after fetching metadata
                                    displayPerformerDetails(performer.name);
                                } catch (error) {
                                    console.error('Error fetching metadata:', error);
                                    alert('Failed to fetch metadata.');
                                }
                                contextMenu.style.display = 'none';
                            };

                            const resetMetadataButton = document.getElementById('reset-metadata-button');
                            resetMetadataButton.onclick = async () => {
                                if (confirm(`Are you sure you want to reset all metadata for ${performer.name}? This cannot be undone.`)) {
                                    try {
                                        const response = await fetch(`/api/performers/${performer.name}/reset-metadata`, {
                                            method: 'POST',
                                        });
                                        if (!response.ok) {
                                            throw new Error(`HTTP error! status: ${response.status}`);
                                        }
                                        const result = await response.json();
                                        console.log(result.message);
                                        alert(result.message);
                                        // Refresh performer details after resetting metadata
                                        displayPerformerDetails(performer.name);
                                    } catch (error) {
                                        console.error('Error resetting metadata:', error);
                                        alert('Failed to reset metadata.');
                                    }
                                }
                                contextMenu.style.display = 'none';
                            };

                            const resetPreviewsButton = document.getElementById('reset-previews-button');
                            resetPreviewsButton.onclick = async () => {
                                try {
                                    const response = await fetch(`/api/performers/${performer.name}/reset-previews`, {
                                        method: 'POST',
                                    });
                                    if (!response.ok) {
                                        throw new Error(`HTTP error! status: ${response.status}`);
                                    }
                                    const result = await response.json();
                                    console.log(result.message);
                                    alert(result.message);
                                    // Refresh performer details after resetting previews
                                    displayPerformerDetails(performer.name);
                                } catch (error) {
                                    console.error('Error resetting previews:', error);
                                    alert('Failed to reset previews.');
                                }
                                contextMenu.style.display = 'none';
                            };
                        });

                        thumbnailNav.appendChild(thumbItem);
                    });

                    // Hide context menu when clicking anywhere else
                    document.addEventListener('click', () => {
                        document.getElementById('context-menu').style.display = 'none';
                    });

                } else {
                    performerCarousel.textContent = 'No previews available.';
                }

                // Populate all tabs
                renderPerformerProfile(performer, performerProfileContent);
                renderPerformerScenes(performer, performerScenesContent);
                renderPerformerAppearance(performer, performerAppearanceContent);
                renderPerformerTags(performer, performerTagsContent);
                renderPerformerBios(performer, performerBiosContent);
                renderPerformerOtherInfo(performer, performerOtherInfoContent);

                // Activate default tab
                performerDetailsPanel.querySelector('.tab-buttons .tab-button.active').click();
            })
            .catch(error => {
                console.error('Error fetching performer details:', error);
                performerProfileContent.textContent = 'Failed to load performer profile.';
                performerScenesContent.textContent = 'Failed to load performer scenes.';
            });
    };

    // Helper function to render key-value pairs from an object
    const renderKeyValuePairs = (data, containerDiv, iconMap) => {
        if (!data || Object.keys(data).length === 0) {
            containerDiv.innerHTML = '<p>No data available.</p>';
            return;
        }
        let html = '<div class="profile-details-grid">';
        for (const key in data) {
            if (data[key] && data[key] !== 'Undefined' && data[key] !== null && data[key] !== 0) {
                const icon = iconMap && iconMap[key] ? iconMap[key] : ''
                html += `
                    <div class="detail-item">
                        <span class="detail-icon">${icon}</span>
                        <span class="detail-label">${key.replace(/_/g, ' ')}:</span>
                        <span class="detail-value">${data[key]}</span>
                    </div>
                `;
            }
        }
        html += '</div>';
        containerDiv.innerHTML = html;
    };

    // Helper function to render a list of items (e.g., tags, external links)
    const renderListItems = (items, containerDiv, isLink = false) => {
        if (!items || items.length === 0) {
            containerDiv.innerHTML = '<p>No data available.</p>';
            return;
        }
        let html = '<ul>';
        items.forEach(item => {
            if (isLink) {
                html += `<li><a href="${item.url}" target="_blank">${item.text || item.url}</a></li>`;
            } else if (typeof item === 'object') {
                // For objects like platform_views, platform_video_counts
                for (const key in item) {
                    if (item[key] && item[key] !== 'Undefined' && item[key] !== null && item[key] !== 0) {
                        html += `<li><strong>${key.replace(/_/g, ' ')}:</strong> ${item[key]}</li>`;
                    }
                }
            } else {
                html += `<li>${item}</li>`;
            }
        });
        html += '</ul>';
        containerDiv.innerHTML = html;
    };

    const renderPerformerProfile = (performer, contentDiv) => {
        const isZoo = performer.zoo === 'true';

        let profileHtml = `
            <h3>${performer.name}</h3>
            <div class="profile-details-grid">
                <div class="detail-item">
                    <span class="detail-icon">👥</span>
                    <span class="detail-label">Aliases:</span>
                    <span class="detail-value">${performer.aliases || 'N/A'}</span>
                </div>
                <div class="detail-item">
                    <span class="detail-icon">🎬</span>
                    <span class="detail-label">Scene Count:</span>
                    <span class="detail-value">${performer.scene_count}</span>
                </div>
                <div class="detail-item">
                    <span class="detail-icon">🐾</span>
                    <span class="detail-label">Zoo:</span>
                    <div class="zoo-toggle-switch ${isZoo ? 'active' : ''}" id="zoo-toggle">
                        <div class="toggle-slider"></div>
                    </div>
                </div>
        `;

        const otherFields = [
            { label: 'Feature Dancer', key: 'feature_dancer', icon: '⭐' },
            { label: 'Date of Birth', key: 'date_of_birth', icon: '🎂' },
            { label: 'Age', key: 'age', icon: '🔞' },
            { label: 'Astrological Sign', key: 'astrological_sign', icon: '♈' },
            { label: 'Career Status', key: 'career_status', icon: '📈' },
            { label: 'Career Start', key: 'career_start', icon: '🏁' },
            { label: 'Career End', key: 'career_end', icon: '🛑' },
            { label: 'Place of Birth', key: 'place_of_birth', icon: '🌍' },
            { label: 'Nationality', key: 'nationality', icon: '🏳️' },
            { label: 'Rank', key: 'rank', icon: '🏆' },
            { label: 'Country', key: 'country', icon: '🗺️' },
        ];

        otherFields.forEach(field => {
            if (performer[field.key] && performer[field.key] !== 'Undefined' && performer[field.key] !== null && performer[field.key] !== 0) {
                profileHtml += `
                    <div class="detail-item">
                        <span class="detail-icon">${field.icon}</span>
                        <span class="detail-label">${field.label}:</span>
                        <span class="detail-value">${performer[field.key]}</span>
                    </div>
                `;
            }
        });

        profileHtml += `</div>`;

        contentDiv.innerHTML = profileHtml;

        const zooToggle = contentDiv.querySelector('#zoo-toggle');
        zooToggle.addEventListener('click', async () => {
            const newZooStatus = !(performer.zoo === 'true');
            try {
                const response = await fetch(`/api/performers/${performer.name}/set-zoo`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ zoo: newZooStatus.toString() }),
                });
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                                    performer.zoo = newZooStatus.toString(); // Update local performer object
                                    zooToggle.classList.toggle('active'); // Update UI
                
                                    // Find the performer in allPerformers and update their zoo status
                                    const performerInAllPerformers = allPerformers.find(p => p.name === performer.name);
                                    if (performerInAllPerformers) {
                                        performerInAllPerformers.zoo = newZooStatus.toString();
                                    }            } catch (error) {
                console.error('Error updating zoo status:', error);
                alert('Failed to update zoo status.');
            }
        });
    };

    const renderPerformerScenes = (performer, contentDiv) => {
        contentDiv.innerHTML = '';
        const associatedVideos = allVideos.filter(video => video.Performers && Array.isArray(video.Performers) && video.Performers.includes(performer.name));
        if (associatedVideos.length === 0) {
            contentDiv.textContent = 'No scenes found for this performer.';
        } else {
            associatedVideos.forEach(video => {
                const videoItem = document.createElement('div');
                videoItem.className = 'video-item';
                videoItem.innerHTML = `
                    <img src="${video.Thumbnail}" alt="${video.Name}" class="video-thumbnail">
                    <div class="video-title-container">
                        <p class="video-title">${video.Name}</p>
                    </div>
                `;
                // Add event listener to open video modal if needed
                contentDiv.appendChild(videoItem);
            });
        }
    };

    const renderPerformerTags = (performer, contentDiv) => {
        if (!performer.tags || performer.tags.length === 0) {
            contentDiv.innerHTML = '<p>No tags available.</p>';
            return;
        }
        let html = '<div class="tags-container">';
        performer.tags.forEach(tag => {
            html += `<span class="tag-badge">#${tag}</span>`;
        });
        html += '</div>';
        contentDiv.innerHTML = html;
    };

    const renderPerformerBios = (performer, contentDiv) => {
        if (!performer.bios || Object.keys(performer.bios).length === 0) {
            contentDiv.innerHTML = '<p>No bios available.</p>';
            return;
        }
        let html = '<div class="bios-container">';
        for (const key in performer.bios) {
            if (performer.bios[key]) {
                html += `
                    <div class="bio-card">
                        <h4>${key.replace(/_/g, ' ')}</h4>
                        <p>${performer.bios[key]}</p>
                    </div>
                `;
            }
        }
        html += '</div>';
        contentDiv.innerHTML = html;
    };



    const renderPerformerAppearance = (performer, contentDiv) => {
        const appearanceIconMap = {
            "ethnicity": "🌍",
            "boobs": "🍈",
            "bust": "📏",
            "cup": "☕",
            "bra": "👙",
            "waist": "👖",
            "hip": "💃",
            "butt": "🍑",
            "height": "📏",
            "weight": "⚖️",
            "hair_color": "ፀጉር",
            "eye_color": "👁️",
            "piercings": "💍",
            "piercing_locations": "📍",
            "tattoos": "🎨",
            "tattoo_locations": "📍",
            "shoe_size": "👟",
            "body_type": "💪",
            "underarm_hair": "🌿",
            "pubic_hair": "🌿"
        };
        renderKeyValuePairs(performer.appearance, contentDiv, appearanceIconMap);
    };

        const renderPerformerOtherInfo = (performer, contentDiv) => {

            const renderOtherInfoSection = (title, icon, data) => {

                if (!data || (typeof data === 'object' && Object.keys(data).length === 0) || (Array.isArray(data) && data.length === 0)) {

                    return '';

                }

    

                let content = '';

                if (Array.isArray(data)) { // For external_links

                    content = '<ul>';

                    data.forEach(item => {

                        if (item.url) {

                            content += `<li><a href="${item.url}" target="_blank">${item.text || item.url}</a></li>`;

                        }

                    });

                    content += '</ul>';

                } else if (typeof data === 'object') { // For performances, social_media, etc.

                    content = '<ul>';

                    for (const key in data) {

                        if (data[key] && data[key] !== 'Undefined' && data[key] !== null && data[key] !== 0) {

                            let value = data[key];

                            if (key.toLowerCase().includes('url') || key.toLowerCase().includes('link') || (typeof value === 'string' && value.startsWith('http'))) {

                                value = `<a href="${value}" target="_blank">${value}</a>`;

                            }

                            content += `<li><strong>${key.replace(/_/g, ' ')}:</strong> ${value}</li>`;

                        }

                    }

                    content += '</ul>';

                } else { // For subscribers, rating, etc.

                    content = `<p>${data}</p>`;

                }

    

                return `

                    <div class="info-card">

                        <div class="info-card-header">

                            <span class="info-card-icon">${icon}</span>

                            <h4>${title}</h4>

                        </div>

                        <div class="info-card-content">

                            ${content}

                        </div>

                    </div>

                `;

            };

    

            let html = '<div class="other-info-container">';

            html += renderOtherInfoSection('Performances', '💃', performer.performances);

            html += renderOtherInfoSection('Social Media', '📱', performer.social_media);

            html += renderOtherInfoSection('Platform Views', '👀', performer.platform_views);

            html += renderOtherInfoSection('Platform Video Counts', '📹', performer.platform_video_counts);

            html += renderOtherInfoSection('Platform Profile Counts', '👥', performer.platform_profile_counts);

            html += renderOtherInfoSection('Subscribers', '📈', performer.subscribers);

            html += renderOtherInfoSection('Rating', '⭐', performer.rating);

            html += renderOtherInfoSection('Total Views', '🔥', performer.total_views);

            html += renderOtherInfoSection('Total Video Count', '🎥', performer.total_video_count);

            html += renderOtherInfoSection('Total Platform Hits', '💥', performer.total_platform_hits);

            html += renderOtherInfoSection('External Links', '🔗', performer.external_links);

            html += '</div>';

    

            contentDiv.innerHTML = html;

        };

    const displayLogs = () => {
        const logOutput = document.getElementById('current-logs').querySelector('pre');
        logOutput.textContent = 'Loading current logs...';

        fetch('/api/logs/current')
            .then(response => {
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                return response.text();
            })
            .then(logContent => {
                logOutput.textContent = logContent;
            })
            .catch(error => {
                console.error('Error fetching current logs:', error);
                logOutput.textContent = 'Failed to load current logs.';
            });
    };

    const displayPreviousLogs = () => {
        const previousLogsPane = document.getElementById('previous-logs');
        previousLogsPane.innerHTML = '<h3>Previous Logs</h3><div id="previous-logs-list">Loading previous logs...</div><pre id="previous-log-content"></pre>';
        const previousLogsList = document.getElementById('previous-logs-list');
        const previousLogContent = document.getElementById('previous-log-content');

        fetch('/api/logs/previous')
            .then(response => {
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                return response.json();
            })
            .then(logFiles => {
                previousLogsList.innerHTML = '';
                if (logFiles.length === 0) {
                    previousLogsList.textContent = 'No previous log files found.';
                    return;
                }

                const ul = document.createElement('ul');
                logFiles.forEach(fileName => {
                    const li = document.createElement('li');
                    const a = document.createElement('a');
                    a.href = '#';
                    a.textContent = fileName;
                    a.addEventListener('click', (e) => {
                        e.preventDefault();
                        previousLogContent.textContent = `Loading ${fileName}...`;
                        fetch(`/api/logs/previous/${fileName}`)
                            .then(response => {
                                if (!response.ok) {
                                    throw new Error(`HTTP error! status: ${response.status}`);
                                }
                                return response.text();
                            })
                            .then(content => {
                                previousLogContent.textContent = content;
                            })
                            .catch(error => {
                                console.error(`Error fetching log file ${fileName}:`, error);
                                previousLogContent.textContent = `Failed to load log file ${fileName}.`;
                            });
                    });
                    li.appendChild(a);
                    ul.appendChild(li);
                });
                previousLogsList.appendChild(ul);
            })
            .catch(error => {
                console.error('Error fetching previous log files:', error);
                previousLogsList.textContent = 'Failed to load previous log files.';
            });
    };

    const handleLogsTabs = () => {
        const logsPage = document.getElementById('logs-page');
        const tabButtons = logsPage.querySelectorAll('.tab-button');
        const tabPanes = logsPage.querySelectorAll('.tab-pane');

        tabButtons.forEach(button => {
            button.addEventListener('click', () => {
                tabButtons.forEach(btn => btn.classList.remove('active'));
                button.classList.add('active');

                tabPanes.forEach(pane => pane.classList.add('hidden'));
                const targetTab = document.getElementById(button.dataset.tab);
                if (targetTab) {
                    targetTab.classList.remove('hidden');
                }

                if (button.dataset.tab === 'current-logs') {
                    displayLogs();
                } else if (button.dataset.tab === 'previous-logs') {
                    displayPreviousLogs();
                }
            });
        });

        // Activate default tab
        const defaultTabButton = logsPage.querySelector('.tab-button.active');
        if (defaultTabButton) {
            defaultTabButton.click();
        }
    };

    closeButton.addEventListener('click', () => {
        closeDetailsPanelAndPauseVideos();
    });
    window.addEventListener('click', (event) => {
        if (event.target == modal) {
            closeDetailsPanelAndPauseVideos();
        }
    });

    const fetchAndRenderGrid = () => {
        videoGrid.innerHTML = ''; // Clear grid
        fetch('/api/videos')
            .then(response => response.json())
            .then(videos => {
                allVideos = videos; // Store all videos globally
                videos.forEach(video => {
                    const videoItem = document.createElement('div');
                    videoItem.className = 'video-item';

                    const thumbnail = document.createElement('img');
                    thumbnail.src = video.thumbnail;
                    thumbnail.className = 'video-thumbnail';

                    const titleContainer = document.createElement('div');
                    titleContainer.className = 'video-title-container';

                    const videoTitle = document.createElement('p');
                    videoTitle.className = 'video-title';
                    videoTitle.textContent = video.name;

                    const videoMetadata = document.createElement('div');
                    videoMetadata.className = 'video-metadata';

                    if (video.metadata && video.metadata.format) {
                        const duration = parseFloat(video.metadata.format.duration);
                        const formattedDuration = new Date(duration * 1000).toISOString().substr(11, 8);
                        const videoStream = video.metadata.streams.find(s => s.codec_type === 'video');
                        const resolution = videoStream ? `${videoStream.width}x${videoStream.height}` : 'N/A';

                        const durationEl = document.createElement('span');
                        durationEl.textContent = formattedDuration;
                        const resolutionEl = document.createElement('span');
                        resolutionEl.textContent = resolution;

                        videoMetadata.appendChild(durationEl);
                        videoMetadata.appendChild(resolutionEl);
                    }

                    titleContainer.appendChild(videoTitle);
                    titleContainer.appendChild(videoMetadata);

                    videoItem.appendChild(thumbnail);
                    videoItem.appendChild(titleContainer);
                    videoGrid.appendChild(videoItem);

                    videoItem.addEventListener('click', () => {
                        openModal(video);
                    });
                });
            })
            .catch(error => {
                console.error('Error fetching videos:', error);
                const errorMsg = document.createElement('p');
                errorMsg.textContent = 'Could not load videos.';
                videoGrid.appendChild(errorMsg);
            });
    };

    // Initial route handling and event listener
    window.addEventListener('hashchange', router);
    router();

    // Update Performer Previews Task
    const updatePerformerPreviewsButton = document.getElementById('update-performer-previews-button');
    const updatePerformerPreviewsStatus = document.getElementById('update-performer-previews-status');

    if (updatePerformerPreviewsButton) {
        updatePerformerPreviewsButton.addEventListener('click', async () => {
            updatePerformerPreviewsStatus.textContent = 'Task started...';
            try {
                const response = await fetch('/api/tasks/update-performer-previews', {
                    method: 'POST',
                });
                const result = await response.json();
                updatePerformerPreviewsStatus.textContent = `Task status: ${result.message}`;
                // Optionally, refresh performers page after task completion
                // handlePerformersPage();
            } catch (error) {
                console.error('Error starting update performer previews task:', error);
                updatePerformerPreviewsStatus.textContent = 'Task failed to start.';
            }
        });
    }

    const refetchAllPerformerMetadataButton = document.getElementById('refetch-all-performer-metadata-button');
    const refetchAllPerformerMetadataStatus = document.getElementById('refetch-all-performer-metadata-status');

    if (refetchAllPerformerMetadataButton) {
        refetchAllPerformerMetadataButton.addEventListener('click', async () => {
            refetchAllPerformerMetadataStatus.textContent = 'Task started...';
            try {
                const response = await fetch('/api/tasks/refetch-all-performer-metadata', {
                    method: 'POST',
                });
                const result = await response.json();
                refetchAllPerformerMetadataStatus.textContent = `Task status: ${result.message}`;
                // Optionally, refresh performers page after task completion
                // handlePerformersPage();
            } catch (error) {
                console.error('Error starting re-fetch all performer metadata task:', error);
                refetchAllPerformerMetadataStatus.textContent = 'Task failed to start.';
            }
        });
    }

    // Chat functionality
    const chatContainer = document.getElementById('chat-container');
    const chatToggleButton = document.getElementById('chat-toggle-button');
    const chatHeader = document.getElementById('chat-header');
    const chatMessages = document.getElementById('chat-messages');
    const chatInput = document.getElementById('chat-input');
    const chatSendButton = document.getElementById('chat-send-button');

    const openChat = () => {
        chatContainer.classList.remove('hidden');
        chatToggleButton.classList.add('hidden');
    };

    const closeChat = () => {
        chatContainer.classList.add('hidden');
        chatToggleButton.classList.remove('hidden');
    };

    chatToggleButton.addEventListener('click', openChat);
    chatHeader.addEventListener('click', closeChat);

    const addMessage = (message, sender) => {
        const messageElement = document.createElement('div');
        messageElement.className = `chat-message ${sender}-message`;
        messageElement.innerHTML = `<div class="message-bubble">${message}</div>`;
        chatMessages.appendChild(messageElement);
        chatMessages.scrollTop = chatMessages.scrollHeight;
    };

    addMessage('Hello! How can I help you organize your videos today?', 'ai');

    const sendMessage = () => {
        const message = chatInput.value.trim();
        if (message === '') return;

        addMessage(message, 'user');
        chatInput.value = '';

        // Send message to backend and get response
        fetch('/api/chat', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ message }),
        })
        .then(response => response.json())
        .then(data => {
            addMessage(data.reply, 'ai');
        });
    };

    chatSendButton.addEventListener('click', sendMessage);
    chatInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            sendMessage();
        }
    });

    // Close details panel when clicking on the navbar
    const navbar = document.querySelector('.navbar');
    if (navbar) {
        navbar.addEventListener('click', (event) => {
            // Prevent closing if a navbar link is clicked (which would navigate)
            if (event.target.tagName === 'A') {
                return;
            }
            closeDetailsPanelAndPauseVideos();
        });
    }

    // Close details panel when clicking on empty space in performer-wall
    const performerWall = document.getElementById('performer-wall');
    if (performerWall) {
        performerWall.addEventListener('click', (event) => {
            // Only close if the click target is the performer-wall itself, not a child element
            if (event.target === performerWall) {
                closeDetailsPanelAndPauseVideos();
            }
        });
    }
});