document.addEventListener('DOMContentLoaded', function() {
    // Create lightbox elements
    const lightboxOverlay = document.createElement('div');
    lightboxOverlay.className = 'lightbox-overlay';
    
    const lightboxContent = document.createElement('div');
    lightboxContent.className = 'lightbox-content';
    
    const lightboxImg = document.createElement('img');
    const lightboxCaption = document.createElement('p');
    lightboxCaption.className = 'lightbox-caption';
    
    lightboxContent.appendChild(lightboxImg);
    lightboxContent.appendChild(lightboxCaption);
    lightboxOverlay.appendChild(lightboxContent);
    document.body.appendChild(lightboxOverlay);

    // Add click handlers to all lightbox links
    document.querySelectorAll('a.lightbox').forEach(link => {
        link.addEventListener('click', function(e) {
            e.preventDefault();
            lightboxImg.src = this.href;
            const caption = this.querySelector('.photo-caption').textContent;
            lightboxCaption.textContent = caption;
            lightboxOverlay.classList.add('active');
        });
    });

    // Close lightbox when clicking anywhere except the caption
    lightboxOverlay.addEventListener('click', function(e) {
        if (!e.target.classList.contains('lightbox-caption')) {
            lightboxOverlay.classList.remove('active');
        }
    });

    // Close lightbox with escape key
    document.addEventListener('keydown', function(e) {
        if (e.key === 'Escape') {
            lightboxOverlay.classList.remove('active');
        }
    });

    // Initialize code blocks
    document.querySelectorAll('.code-block-wrapper').forEach(wrapper => {
        const pre = wrapper.querySelector('pre');
        
        // Check if the content is taller than max-height
        if (pre.scrollHeight > pre.clientHeight) {
            wrapper.classList.add('collapsible');
        }
        
        // Add click handler for expand button
        const expandBtn = wrapper.querySelector('.expand-code');
        if (expandBtn) {
            expandBtn.addEventListener('click', () => {
                wrapper.classList.toggle('expanded');
            });
        }
    });

    // Progressive image loading
    const images = document.querySelectorAll('.photo-item img');
    
    if ('loading' in HTMLImageElement.prototype) {
        // Browser supports native lazy loading
        images.forEach(img => {
            img.loading = 'lazy';
        });
    } else {
        // Fallback for browsers that don't support lazy loading
        const imageObserver = new IntersectionObserver((entries, observer) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    const img = entry.target;
                    img.src = img.dataset.src;
                    observer.unobserve(img);
                }
            });
        });

        images.forEach(img => {
            img.dataset.src = img.src;
            img.src = ''; // Clear src to prevent immediate loading
            imageObserver.observe(img);
        });
    }
});
