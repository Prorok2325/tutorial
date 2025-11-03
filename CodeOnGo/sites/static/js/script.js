// Анимации и интерактивность для крипто-платформы
document.addEventListener('DOMContentLoaded', function() {
    // Анимация появления элементов
    const observerOptions = {
        threshold: 0.1,
        rootMargin: '0px 0px -50px 0px'
    };

    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.style.opacity = '1';
                entry.target.style.transform = 'translateY(0)';
            }
        });
    }, observerOptions);

    // Применяем анимацию к карточкам заданий
    const taskCards = document.querySelectorAll('.task-card');
    const steps = document.querySelectorAll('.step');
    
    [...taskCards, ...steps].forEach(item => {
        item.style.opacity = '0';
        item.style.transform = 'translateY(30px)';
        item.style.transition = 'opacity 0.6s ease, transform 0.6s ease';
        observer.observe(item);
    });

    // Фильтрация заданий
    const filterTabs = document.querySelectorAll('.filter-tab');
    filterTabs.forEach(tab => {
        tab.addEventListener('click', function() {
            // Убираем активный класс у всех вкладок
            filterTabs.forEach(t => t.classList.remove('active'));
            // Добавляем активный класс текущей вкладке
            this.classList.add('active');
            
            // Здесь можно добавить логику фильтрации
            const filter = this.textContent.toLowerCase();
            filterTasks(filter);
        });
    });

    // Обработка кнопок "Взять задание"
    const taskButtons = document.querySelectorAll('.btn-task:not(.completed)');
    taskButtons.forEach(button => {
        button.addEventListener('click', function() {
            const taskId = this.getAttribute('data-task');
            takeTask(taskId, this);
        });
    });

    // Анимация чисел в статистике
    animateNumbers();
});

function filterTasks(filter) {
    const tasks = document.querySelectorAll('.task-card');
    tasks.forEach(task => {
        if (filter === 'все') {
            task.style.display = 'block';
        } else {
            const category = task.querySelector('.task-category').textContent.toLowerCase();
            if (category.includes(filter)) {
                task.style.display = 'block';
            } else {
                task.style.display = 'none';
            }
        }
    });
}

function takeTask(taskId, button) {
    // Имитация взятия задания
    button.textContent = 'Задание принято...';
    button.disabled = true;
    button.style.background = '#888';
    
    setTimeout(() => {
        button.textContent = 'Ожидает проверки';
        // В реальном приложении здесь был бы AJAX запрос к серверу
    }, 1500);
}

function animateNumbers() {
    const statValues = document.querySelectorAll('.stat-value');
    statValues.forEach(stat => {
        const finalValue = stat.textContent;
        let currentValue = 0;
        const increment = parseInt(finalValue) / 100;
        const timer = setInterval(() => {
            currentValue += increment;
            if (currentValue >= parseInt(finalValue)) {
                stat.textContent = finalValue;
                clearInterval(timer);
            } else {
                stat.textContent = Math.floor(currentValue) + (finalValue.includes('+') ? '+' : '');
            }
        }, 20);
    });
}

// Обновление времени в реальном времени
function updateTime() {
    const now = new Date();
    const timeElement = document.querySelector('.current-time');
    if (timeElement) {
        timeElement.textContent = now.toLocaleTimeString();
    }
}

setInterval(updateTime, 1000);

// Эффект частиц для фона (упрощенная версия)
function createParticles() {
    const hero = document.querySelector('.hero');
    if (!hero) return;
    
    for (let i = 0; i < 20; i++) {
        const particle = document.createElement('div');
        particle.style.position = 'absolute';
        particle.style.width = '2px';
        particle.style.height = '2px';
        particle.style.background = 'var(--primary)';
        particle.style.borderRadius = '50%';
        particle.style.left = Math.random() * 100 + '%';
        particle.style.top = Math.random() * 100 + '%';
        particle.style.opacity = '0.3';
        particle.style.animation = `float ${3 + Math.random() * 7}s infinite ease-in-out`;
        hero.appendChild(particle);
    }
}

// Добавляем CSS анимацию для частиц
const style = document.createElement('style');
style.textContent = `
    @keyframes float {
        0%, 100% { transform: translate(0, 0) rotate(0deg); }
        25% { transform: translate(10px, -10px) rotate(90deg); }
        50% { transform: translate(0, -20px) rotate(180deg); }
        75% { transform: translate(-10px, -10px) rotate(270deg); }
    }
`;
document.head.appendChild(style);

createParticles();