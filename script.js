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

    // Handle image orientation
    const photoGrid = document.querySelector('.photo-grid');
    if (photoGrid) {
        const photoItems = photoGrid.querySelectorAll('.photo-item img');
        
        photoItems.forEach(img => {
            if (img.complete) {
                setImageOrientation(img);
            } else {
                img.addEventListener('load', () => setImageOrientation(img));
            }
        });
    }
});
