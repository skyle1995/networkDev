const VERSION = '2.10.1';
const layuicss = `https://unpkg.com/layui@${VERSION}/dist/css/layui.css`;
const layuijs = `https://unpkg.com/layui@${VERSION}/dist/layui.js`;
const rootPath = (function (src) {
  src = (document.currentScript && document.currentScript.tagName.toUpperCase() === 'SCRIPT') ? document.currentScript.src : document.scripts[document.scripts.length - 1].src;
  return src.substring(0, src.lastIndexOf('/') + 1);
})();

const app = document.querySelector('#app')

addLink({ href: layuicss }).then(() => {
  app.style.display = 'block';
});

addLink({ id: 'layui_theme_css', href: `./static/src/layui-theme-dark-selector.css` });

// TODO 弃用，下个版本只支持选择器模式
//addLink({ id: 'layui_theme_css', href: `${rootPath}dist/layui-theme-dark.css` });

loadScript(layuijs, function () {
  layui
    .config({
      base: './static/lib/',
    })
    .extend({
      drawer: 'drawer/drawer',
    });
  layui.use(['drawer', 'colorMode'], function () {
    const { $, element, form, layer, util, dropdown, drawer, colorMode } = layui;

    const APPERANCE_KEY = 'layui-theme-demo-prefer-dark';

    const theme = colorMode.init({
      selector: 'html',
      attribute: 'class',
      initialValue: 'dark',
      modes: {
        light: '',
        dark: 'dark',
      },
      storageKey: APPERANCE_KEY,
      onChanged(mode, defaultHandler) {
        const isAppearanceTransition = document.startViewTransition && !window.matchMedia(`(prefers-reduced-motion: reduce)`).matches;
        const isDark = mode === 'dark';

        $('#change-theme').attr('class', `layui-icon layui-icon-${isDark ? 'moon' : 'light'}`);

        if (!isAppearanceTransition) {
          defaultHandler();
        } else {
          rippleViewTransition(isDark, function () {
            defaultHandler();
          });
        }
      },
    });

    routerTo({path: location.hash.slice(1) || 'dashboard'});

    dropdown.render({
      elem: '#change-theme',
      align: 'center',
      data: [
        {
          title: '深色模式',
          id: 'dark',
          icon: 'layui-icon-moon',
        },
        {
          title: '浅色模式',
          id: 'light',
          icon: 'layui-icon-light',
        },
        {
          title: '跟随系统',
          id: 'auto',
          icon: 'layui-icon-console',
        },
      ],
      templet(d) {
        return `
                <span style="display: flex;">
                  <i class="layui-icon ${d.icon}" style="margin-right: 8px"></i>
                  ${d.title}
                </span>`.trim();
      },
      click(obj) {
        const { id: mode } = obj;
        theme.setMode(mode);
      },
    });

    util.event('lay-header-event', {
      menuLeft() {
        $('body').toggleClass('collapse');
      },
      menuRight() {
        drawer.open({
          area: '600px',
          url: './static/tpl/theme.html',
          hideOnClose: true,
          id: 'drawer-theme-tpl',
          shade: 0.01,
        });
      },
    });

    element.on('nav(nav-side)', function (elem) {
      var path = elem.data('path');
      if (path) {
        routerTo({path});
        if ($(window).width() <= 768) {
          $('body').toggleClass('collapse', false);
        }
      }
    });

    $('#layuiv').text(layui.v);

    /*
     * 后台通用脚本
     * 说明：统一处理全局的退出登录逻辑，遵循后端 jsonResponse 的返回格式：
     * code: 0 表示成功，非0表示失败
     * msg: 提示信息
     * data: 业务数据
     */
    
    // 绑定退出登录按钮事件（箭头函数写法）
    const bindLogout = () => {
      const btn = document.getElementById('logout-btn');
      if (!btn) return;
      btn.addEventListener('click', (e) => {
        e.preventDefault();
        handleLogout();
      });
    };
    
    // 执行退出登录（箭头函数写法）
    // 功能：弹出确认框 -> 显示加载层 -> 调用 /admin/logout -> 依据 code===0 判断
    const handleLogout = () => {
      layer.confirm('确定要退出登录吗？', {
        icon: 3,
        title: '提示'
      }, (index) => {
        layer.close(index);
        
        // 显示加载层
        const loadIndex = layer.load(2, {
          content: '正在退出登录...'
        });
        
        // 调用登出接口
        fetch('/admin/logout', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'X-Requested-With': 'XMLHttpRequest'
          }
        })
        .then(response => response.json())
        .then(data => {
          layer.close(loadIndex);
          const ok = data && data.code === 0;
          const msg = (data && (data.msg || data.message)) || (ok ? '退出登录成功' : '退出登录失败');
          if (ok) {
            layer.msg(msg, {
              icon: 1,
              time: 1000
            }, () => {
              // 跳转到登录页或后端返回的地址
              const redirect = (data && data.data && data.data.redirect) || '/admin/login';
              window.location.href = redirect;
            });
          } else {
            layer.msg(msg, { icon: 2 });
          }
        })
        .catch(error => {
          layer.close(loadIndex);
          console.error('登出请求失败:', error);
          layer.msg('网络错误，请重试', { icon: 2 });
        });
      });
    };
    
    // 页面就绪后绑定事件（箭头函数写法）
    (() => {
      if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', bindLogout);
      } else {
        bindLogout();
      }
    })();

    // 刷新页面功能处理
    const handleRefresh = () => {
      layer.confirm('确定要刷新内容吗？', {
        icon: 3,
        title: '提示'
      }, (index) => {
        layer.close(index);
        
        // 获取当前hash值，确定当前页面路径
        let currentPath = window.location.hash.replace('#', '') || 'dashboard';
        
        // 显示加载层
        const loadIndex = layer.load(2, {
          content: '正在刷新...'
        });
        
        // 延迟一下再刷新内容，让用户看到加载效果
        setTimeout(() => {
          // 重新加载当前内容页面
          routerTo({ path: currentPath });
          layer.close(loadIndex);
        }, 500);
      });
    };

    // 绑定刷新按钮点击事件
    $('#refresh-btn').on('click', handleRefresh);

    function routerTo({
      elem = '#router-view',
      path = 'dashboard',
      prefix = 'admin/', //路由前缀
      suffix = '', //路由后缀
    } = {}) {
      var routerView = $(elem);
      var url = prefix + path + suffix;

      var loadTimer = setTimeout(() => {
        layer.load(2);
      }, 100);

      history.replaceState({}, '', `#${path}`); // 因为并没有处理路由
      routerView.attr('src', url)
      routerView.off('load').on('load',function(){
        element.render();
        form.render();
        clearTimeout(loadTimer);
        layer.closeLast('loading');
      })

      // 选中, 展开菜单
      $('#ws-nav-side')
        .find("[data-path='" + path + "']")
        .parent('dd')
        .addClass('layui-this')
        .closest('.layui-nav-item')
        .addClass('layui-nav-itemed');
    }

  });
});

function rippleViewTransition(isDark, callback) {
  // 移植自 https://github.com/vuejs/vitepress/pull/2347
  // 支持 Chrome 111+

  // 兼容 jQuery 3 下隐式 event 全局对象不可用的问题
  if (!window.event) {
    window.event = new MouseEvent('click', {
      clientX: document.documentElement.clientWidth,
      clientY: 60,
    });
  }

  const x = event.clientX;
  const y = event.clientY;
  const endRadius = Math.hypot(Math.max(x, innerWidth - x), Math.max(y, innerHeight - y));
  const transition = document.startViewTransition(function () {
    callback && callback();
  });
  transition.ready.then(function () {
    var clipPath = [`circle(0px at ${x}px ${y}px)`, `circle(${endRadius}px at ${x}px ${y}px)`];
    document.documentElement.animate(
      {
        clipPath: isDark ? clipPath : [...clipPath].reverse(),
      },
      {
        duration: 300,
        easing: 'ease-in',
        pseudoElement: isDark ? '::view-transition-new(root)' : '::view-transition-old(root)',
      }
    );
  });
}

function addStyle(id, cssStr) {
  const el = document.getElementById(id) || document.createElement('style');
  if (!el.isConnected) {
    el.type = 'text/css';
    el.id = id;
    document.head.appendChild(el);
  }
  el.textContent = cssStr;
}

function addLink(opt) {
  return new Promise((resolve) => {
    const link = Object.assign(document.createElement('link'), {
      rel: 'stylesheet',
      onload: () => resolve({ ...opt, status: 'success' }),
      onerror: () => resolve({ ...opt, status: 'error' }), // 为了在 Promise.all 的使用场景
      ...opt,
    });
    document.head.appendChild(link);
  });
}

function loadScript(url, callback) {
  const script = document.createElement('script');
  script.type = 'text/javascript';
  script.async = 'async';
  script.src = url;
  document.body.appendChild(script);
  if (script.readyState) {
    script.onreadystatechange = function () {
      if (script.readyState == 'complete' || script.readyState == 'loaded') {
        script.onreadystatechange = null;
        callback && callback();
      }
    };
  } else {
    script.onload = function () {
      callback && callback();
    };
  }
}
