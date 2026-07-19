// @ts-nocheck
    function setActiveLink(element) {
      document.querySelectorAll('.sidebar-link').forEach(link => {
        link.classList.remove('active');
      });
      element.classList.add('active');
    }

    function copyCode(btn) {
      const codeBlock = btn.nextElementSibling.querySelector('code');
      const text = codeBlock.textContent;
      
      navigator.clipboard.writeText(text).then(() => {
        const originalText = btn.textContent;
        btn.textContent = 'Copied!';
        btn.style.color = '#58a6ff';
        btn.style.borderColor = '#58a6ff';
        
        setTimeout(() => {
          btn.textContent = originalText;
          btn.style.color = '';
          btn.style.borderColor = '';
        }, 2000);
      });
    }

    // Scroll spy for sidebar links
    window.addEventListener('DOMContentLoaded', () => {
      const observer = new IntersectionObserver(entries => {
        entries.forEach(entry => {
          const id = entry.target.getAttribute('id');
          if (entry.isIntersecting) {
            document.querySelectorAll('.sidebar-link').forEach(link => {
              link.classList.remove('active');
              if (link.getAttribute('href') === `#${id}`) {
                link.classList.add('active');
              }
            });
          }
        });
      }, {
        rootMargin: '-20% 0px -60% 0px'
      });

      document.querySelectorAll('.api-card').forEach(card => {
        observer.observe(card);
      });
    });
