export default function AboutPage() {
  return (
    <div className="max-w-3xl mx-auto">
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-6">关于 SayIt</h1>

        <div className="prose prose-gray max-w-none space-y-6">
          <p className="text-lg text-gray-600 italic">
            晒一个有意思的灵魂
          </p>

          <p className="text-gray-700 leading-relaxed">
            SayIt 是一个仿 Reddit 的社区论坛平台，旨在为用户提供一个自由交流、分享观点和发现有趣内容的空间。
          </p>

          <h2 className="text-xl font-semibold text-gray-900 mt-8">项目背景</h2>
          <p className="text-gray-700 leading-relaxed">
            SayIt 作为一个全栈学习项目，后端采用 Go 语言开发，使用 Gin 框架、GORM、MySQL 和 Redis；
            前端使用 Next.js + TypeScript + Tailwind CSS 构建。项目涵盖了用户认证、帖子系统、社区分类、
            投票机制等核心功能。
          </p>

          <h2 className="text-xl font-semibold text-gray-900 mt-8">技术栈</h2>
          <div className="grid grid-cols-2 gap-4 mt-4">
            <div className="bg-gray-50 rounded-lg p-4">
              <h3 className="font-semibold text-gray-900 mb-2">后端</h3>
              <ul className="text-sm text-gray-600 space-y-1">
                <li>Go + Gin Web 框架</li>
                <li>GORM (MySQL ORM)</li>
                <li>Redis (排行榜/缓存)</li>
                <li>JWT 认证</li>
                <li>雪花算法 (ID 生成)</li>
              </ul>
            </div>
            <div className="bg-gray-50 rounded-lg p-4">
              <h3 className="font-semibold text-gray-900 mb-2">前端</h3>
              <ul className="text-sm text-gray-600 space-y-1">
                <li>Next.js 14 (App Router)</li>
                <li>TypeScript</li>
                <li>Tailwind CSS</li>
                <li>React Context 状态管理</li>
                <li>Lucide Icons</li>
              </ul>
            </div>
          </div>

          <h2 className="text-xl font-semibold text-gray-900 mt-8">核心功能</h2>
          <ul className="text-gray-700 space-y-2">
            <li>用户注册与登录（JWT 认证）</li>
            <li>社区创建与浏览</li>
            <li>帖子发布、查看与投票</li>
            <li>按时间或热度排序</li>
            <li>响应式布局设计</li>
          </ul>
        </div>
      </div>
    </div>
  );
}
