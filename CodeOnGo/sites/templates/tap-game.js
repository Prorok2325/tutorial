class TapGame {
    constructor() {
        this.bitcoin = 0;
        this.perTap = 0.00000001;
        this.perSecond = 0;
        this.multiplier = 1;
        
        this.upgrades = {
            click: { level: 1, cost: 0.00000010, baseCost: 0.00000010 },
            auto: { level: 0, cost: 0.00000100, baseCost: 0.00000100 },
            multiplier: { level: 1, cost: 0.00001000, baseCost: 0.00001000 }
        };
        
        this.achievements = {
            firstBitcoin: false,
            cryptoEnthusiast: false,
            bitcoinMillionaire: false
        };
        
        this.combo = 0;
        this.lastTapTime = 0;
        this.comboTimeout = null;
        
        this.loadGame();
        this.startGameLoop();
        this.setupEventListeners();
    }
    
    setupEventListeners() {
        const tapElement = document.getElementById('bitcoinTap');
        tapElement.addEventListener('click', (e) => this.handleTap(e));
        tapElement.addEventListener('touchstart', (e) => {
            e.preventDefault();
            this.handleTap(e);
        });
    }
    
    handleTap(event) {
        const currentTime = Date.now();
        const tapDelay = currentTime - this.lastTapTime;
        
        // Комбо система
        if (tapDelay < 500) {
            this.combo++;
            this.showCombo();
        } else {
            this.combo = 1;
        }
        
        this.lastTapTime = currentTime;
        
        // Сброс комбо через 1 секунду
        clearTimeout(this.comboTimeout);
        this.comboTimeout = setTimeout(() => {
            this.combo = 0;
            this.hideCombo();
        }, 1000);
        
        // Начисление BTC с учетом комбо
        let tapValue = this.perTap * this.multiplier;
        if (this.combo > 5) {
            tapValue *= (1 + (this.combo - 5) * 0.1); // +10% за каждый комбо выше 5
        }
        
        this.bitcoin += tapValue;
        this.createParticles(event);
        this.updateDisplay();
        this.checkAchievements();
        this.saveGame();
    }
    
    createParticles(event) {
        const tapElement = document.getElementById('bitcoinTap');
        const rect = tapElement.getBoundingClientRect();
        const centerX = rect.left + rect.width / 2;
        const centerY = rect.top + rect.height / 2;
        
        for (let i = 0; i < 5; i++) {
            const particle = document.createElement('div');
            particle.className = 'particle';
            particle.style.left = centerX + 'px';
            particle.style.top = centerY + 'px';
            particle.style.background = i % 2 === 0 ? '#ffd700' : '#f7931a';
            
            document.body.appendChild(particle);
            
            const angle = Math.random() * Math.PI * 2;
            const distance = 50 + Math.random() * 100;
            const duration = 0.5 + Math.random() * 0.5;
            
            particle.style.animation = `floatUp ${duration}s ease-out forwards`;
            particle.style.transform = `translate(${Math.cos(angle) * distance}px, ${Math.sin(angle) * distance}px)`;
            
            setTimeout(() => {
                if (particle.parentNode) {
                    particle.parentNode.removeChild(particle);
                }
            }, duration * 1000);
        }
    }
    
    showCombo() {
        let comboDisplay = document.querySelector('.combo-display');
        if (!comboDisplay) {
            comboDisplay = document.createElement('div');
            comboDisplay.className = 'combo-display';
            document.body.appendChild(comboDisplay);
        }
        
        if (this.combo >= 5) {
            comboDisplay.textContent = `Комбо x${this.combo}! +${((this.combo - 5) * 10)}%`;
            comboDisplay.style.display = 'block';
        }
    }
    
    hideCombo() {
        const comboDisplay = document.querySelector('.combo-display');
        if (comboDisplay) {
            comboDisplay.style.display = 'none';
        }
    }
    
    buyUpgrade(type) {
        const upgrade = this.upgrades[type];
        
        if (this.bitcoin >= upgrade.cost) {
            this.bitcoin -= upgrade.cost;
            
            switch (type) {
                case 'click':
                    upgrade.level++;
                    this.perTap += 0.00000001;
                    upgrade.cost = upgrade.baseCost * Math.pow(1.5, upgrade.level - 1);
                    break;
                    
                case 'auto':
                    upgrade.level++;
                    this.perSecond += 0.00000001;
                    upgrade.cost = upgrade.baseCost * Math.pow(1.8, upgrade.level);
                    break;
                    
                case 'multiplier':
                    upgrade.level++;
                    this.multiplier *= 2;
                    upgrade.cost = upgrade.baseCost * Math.pow(3, upgrade.level - 1);
                    break;
            }
            
            this.updateDisplay();
            this.saveGame();
        }
    }
    
    updateDisplay() {
        document.getElementById('bitcoinCount').textContent = this.bitcoin.toFixed(8);
        document.getElementById('perTap').textContent = (this.perTap * this.multiplier).toFixed(8);
        document.getElementById('perSecond').textContent = (this.perSecond * this.multiplier).toFixed(8);
        
        // Обновление стоимостей улучшений
        document.getElementById('clickCost').textContent = this.upgrades.click.cost.toFixed(8) + ' BTC';
        document.getElementById('autoCost').textContent = this.upgrades.auto.cost.toFixed(8) + ' BTC';
        document.getElementById('multiplierCost').textContent = this.upgrades.multiplier.cost.toFixed(8) + ' BTC';
        
        // Обновление уровней улучшений
        document.getElementById('clickLevel').textContent = this.upgrades.click.level;
        document.getElementById('autoLevel').textContent = this.upgrades.auto.level;
        document.getElementById('multiplierLevel').textContent = this.upgrades.multiplier.level;
        
        // Блокировка улучшений, которые нельзя купить
        const upgrades = document.querySelectorAll('.upgrade-card');
        upgrades[0].className = `upgrade-card ${this.bitcoin >= this.upgrades.click.cost ? '' : 'disabled'}`;
        upgrades[1].className = `upgrade-card ${this.bitcoin >= this.upgrades.auto.cost ? '' : 'disabled'}`;
        upgrades[2].className = `upgrade-card ${this.bitcoin >= this.upgrades.multiplier.cost ? '' : 'disabled'}`;
    }
    
    checkAchievements() {
        if (!this.achievements.firstBitcoin && this.bitcoin >= 0.00000100) {
            this.achievements.firstBitcoin = true;
            document.getElementById('ach1').classList.add('unlocked');
            this.showNotification('🥉 Достижение разблокировано: Первый биткоин!');
        }
        
        if (!this.achievements.cryptoEnthusiast && this.bitcoin >= 0.00010000) {
            this.achievements.cryptoEnthusiast = true;
            document.getElementById('ach2').classList.add('unlocked');
            this.showNotification('🥈 Достижение разблокировано: Крипто-энтузиаст!');
        }
        
        if (!this.achievements.bitcoinMillionaire && this.bitcoin >= 0.01000000) {
            this.achievements.bitcoinMillionaire = true;
            document.getElementById('ach3').classList.add('unlocked');
            this.showNotification('🥇 Достижение разблокировано: Биткоин-миллионер!');
        }
    }
    
    showNotification(message) {
        const notification = document.createElement('div');
        notification.style.cssText = `
            position: fixed;
            top: 100px;
            left: 50%;
            transform: translateX(-50%);
            background: rgba(247, 147, 26, 0.9);
            color: white;
            padding: 1rem 2rem;
            border-radius: 10px;
            font-weight: bold;
            z-index: 1000;
            animation: slideDown 0.3s ease;
        `;
        
        const style = document.createElement('style');
        style.textContent = `
            @keyframes slideDown {
                from { top: 0; opacity: 0; }
                to { top: 100px; opacity: 1; }
            }
        `;
        document.head.appendChild(style);
        
        notification.textContent = message;
        document.body.appendChild(notification);
        
        setTimeout(() => {
            notification.remove();
            style.remove();
        }, 3000);
    }
    
    startGameLoop() {
        setInterval(() => {
            if (this.perSecond > 0) {
                this.bitcoin += this.perSecond * this.multiplier;
                this.updateDisplay();
                this.checkAchievements();
                this.saveGame();
            }
        }, 1000);
    }
    
    saveGame() {
        const gameData = {
            bitcoin: this.bitcoin,
            perTap: this.perTap,
            perSecond: this.perSecond,
            multiplier: this.multiplier,
            upgrades: this.upgrades,
            achievements: this.achievements
        };
        localStorage.setItem('bitcoinTapGame', JSON.stringify(gameData));
    }
    
    loadGame() {
        const saved = localStorage.getItem('bitcoinTapGame');
        if (saved) {
            const gameData = JSON.parse(saved);
            this.bitcoin = gameData.bitcoin || 0;
            this.perTap = gameData.perTap || 0.00000001;
            this.perSecond = gameData.perSecond || 0;
            this.multiplier = gameData.multiplier || 1;
            this.upgrades = gameData.upgrades || this.upgrades;
            this.achievements = gameData.achievements || this.achievements;
            
            // Разблокируем достижения
            if (this.achievements.firstBitcoin) {
                document.getElementById('ach1').classList.add('unlocked');
            }
            if (this.achievements.cryptoEnthusiast) {
                document.getElementById('ach2').classList.add('unlocked');
            }
            if (this.achievements.bitcoinMillionaire) {
                document.getElementById('ach3').classList.add('unlocked');
            }
        }
        this.updateDisplay();
    }
    
    resetGame() {
        if (confirm('Вы уверены, что хотите сбросить весь прогресс?')) {
            this.bitcoin = 0;
            this.perTap = 0.00000001;
            this.perSecond = 0;
            this.multiplier = 1;
            
            this.upgrades = {
                click: { level: 1, cost: 0.00000010, baseCost: 0.00000010 },
                auto: { level: 0, cost: 0.00000100, baseCost: 0.00000100 },
                multiplier: { level: 1, cost: 0.00001000, baseCost: 0.00001000 }
            };
            
            this.achievements = {
                firstBitcoin: false,
                cryptoEnthusiast: false,
                bitcoinMillionaire: false
            };
            
            document.getElementById('ach1').classList.remove('unlocked');
            document.getElementById('ach2').classList.remove('unlocked');
            document.getElementById('ach3').classList.remove('unlocked');
            
            this.updateDisplay();
            this.saveGame();
        }
    }
}

// Глобальные функции для вызова из HTML
let game;

function initGame() {
    game = new TapGame();
}

function buyUpgrade(type) {
    if (game) {
        game.buyUpgrade(type);
    }
}

function resetGame() {
    if (game) {
        game.resetGame();
    }
}

function saveGame() {
    if (game) {
        game.saveGame();
        alert('Игра сохранена!');
    }
}

// Инициализация игры при загрузке страницы
document.addEventListener('DOMContentLoaded', initGame);