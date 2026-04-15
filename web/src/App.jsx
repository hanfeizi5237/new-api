/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { lazy, Suspense, useContext, useMemo } from 'react';
import {
  Navigate,
  Route,
  Routes,
  useLocation,
  useParams,
} from 'react-router-dom';
import Loading from './components/common/ui/Loading';
import { AuthRedirect, PrivateRoute, AdminRoute } from './helpers';
import RegisterForm from './components/auth/RegisterForm';
import LoginForm from './components/auth/LoginForm';
import NotFound from './pages/NotFound';
import Forbidden from './pages/Forbidden';
import { StatusContext } from './context/Status';

import PasswordResetForm from './components/auth/PasswordResetForm';
import PasswordResetConfirm from './components/auth/PasswordResetConfirm';
import OAuth2Callback from './components/auth/OAuth2Callback';
import SetupCheck from './components/layout/SetupCheck';

const Home = lazy(() => import('./pages/Home'));
const Dashboard = lazy(() => import('./pages/Dashboard'));
const About = lazy(() => import('./pages/About'));
const UserAgreement = lazy(() => import('./pages/UserAgreement'));
const PrivacyPolicy = lazy(() => import('./pages/PrivacyPolicy'));
const Setup = lazy(() => import('./pages/Setup'));
const User = lazy(() => import('./pages/User'));
const Setting = lazy(() => import('./pages/Setting'));
const Channel = lazy(() => import('./pages/Channel'));
const Entitlement = lazy(() => import('./pages/Entitlement'));
const Listing = lazy(() => import('./pages/Listing'));
const Token = lazy(() => import('./pages/Token'));
const Order = lazy(() => import('./pages/Order'));
const Redemption = lazy(() => import('./pages/Redemption'));
const Seller = lazy(() => import('./pages/Seller'));
const TopUp = lazy(() => import('./pages/TopUp'));
const Log = lazy(() => import('./pages/Log'));
const Marketplace = lazy(() => import('./pages/Marketplace'));
const Chat = lazy(() => import('./pages/Chat'));
const Chat2Link = lazy(() => import('./pages/Chat2Link'));
const Midjourney = lazy(() => import('./pages/Midjourney'));
const Pricing = lazy(() => import('./pages/Pricing'));
const Task = lazy(() => import('./pages/Task'));
const ModelPage = lazy(() => import('./pages/Model'));
const ModelDeploymentPage = lazy(() => import('./pages/ModelDeployment'));
const Playground = lazy(() => import('./pages/Playground'));
const Subscription = lazy(() => import('./pages/Subscription'));
const PersonalSetting = lazy(
  () => import('./components/settings/PersonalSetting'),
);

function DynamicOAuth2Callback() {
  const { provider } = useParams();
  return <OAuth2Callback type={provider} />;
}

function MarketplaceLegacyRedirect() {
  const location = useLocation();
  return <Navigate to={`/console/marketplace${location.search}`} replace />;
}

function App() {
  const location = useLocation();
  const [statusState] = useContext(StatusContext);
  const renderRoute = (element) => (
    // Lazy route boundaries keep large admin/marketplace modules out of the initial shell bundle.
    <Suspense fallback={<Loading></Loading>} key={location.pathname}>
      {element}
    </Suspense>
  );

  // 获取模型广场权限配置
  const pricingRequireAuth = useMemo(() => {
    const headerNavModulesConfig = statusState?.status?.HeaderNavModules;
    if (headerNavModulesConfig) {
      try {
        const modules = JSON.parse(headerNavModulesConfig);

        // 处理向后兼容性：如果pricing是boolean，默认不需要登录
        if (typeof modules.pricing === 'boolean') {
          return false; // 默认不需要登录鉴权
        }

        // 如果是对象格式，使用requireAuth配置
        return modules.pricing?.requireAuth === true;
      } catch (error) {
        console.error('解析顶栏模块配置失败:', error);
        return false; // 默认不需要登录
      }
    }
    return false; // 默认不需要登录
  }, [statusState?.status?.HeaderNavModules]);

  return (
    <SetupCheck>
      <Routes>
        <Route path='/' element={renderRoute(<Home />)} />
        <Route path='/setup' element={renderRoute(<Setup />)} />
        <Route path='/forbidden' element={<Forbidden />} />
        <Route
          path='/console/models'
          element={
            <AdminRoute>
              {renderRoute(<ModelPage />)}
            </AdminRoute>
          }
        />
        <Route
          path='/console/deployment'
          element={
            <AdminRoute>
              {renderRoute(<ModelDeploymentPage />)}
            </AdminRoute>
          }
        />
        <Route
          path='/console/subscription'
          element={
            <AdminRoute>
              {renderRoute(<Subscription />)}
            </AdminRoute>
          }
        />
        <Route
          path='/console/channel'
          element={
            <AdminRoute>
              {renderRoute(<Channel />)}
            </AdminRoute>
          }
        />
        <Route
          path='/console/token'
          element={
            <PrivateRoute>
              {renderRoute(<Token />)}
            </PrivateRoute>
          }
        />
        <Route
          path='/console/playground'
          element={
            <PrivateRoute>
              {renderRoute(<Playground />)}
            </PrivateRoute>
          }
        />
        <Route
          path='/console/marketplace'
          element={
            <PrivateRoute>
              {renderRoute(<Marketplace />)}
            </PrivateRoute>
          }
        />
        <Route
          path='/console/redemption'
          element={
            <AdminRoute>
              {renderRoute(<Redemption />)}
            </AdminRoute>
          }
        />
        <Route
          path='/console/user'
          element={
            <AdminRoute>
              {renderRoute(<User />)}
            </AdminRoute>
          }
        />
        <Route
          path='/console/sellers'
          element={
            <AdminRoute>
              {renderRoute(<Seller />)}
            </AdminRoute>
          }
        />
        <Route
          path='/console/listings'
          element={
            <AdminRoute>
              {renderRoute(<Listing />)}
            </AdminRoute>
          }
        />
        <Route
          path='/console/orders'
          element={
            <AdminRoute>
              {renderRoute(<Order />)}
            </AdminRoute>
          }
        />
        <Route
          path='/console/entitlements'
          element={
            <AdminRoute>
              {renderRoute(<Entitlement />)}
            </AdminRoute>
          }
        />
        <Route path='/user/reset' element={renderRoute(<PasswordResetConfirm />)} />
        <Route
          path='/login'
          element={
            <Suspense fallback={<Loading></Loading>} key={location.pathname}>
              <AuthRedirect>
                <LoginForm />
              </AuthRedirect>
            </Suspense>
          }
        />
        <Route
          path='/register'
          element={
            <Suspense fallback={<Loading></Loading>} key={location.pathname}>
              <AuthRedirect>
                <RegisterForm />
              </AuthRedirect>
            </Suspense>
          }
        />
        <Route path='/reset' element={renderRoute(<PasswordResetForm />)} />
        <Route path='/oauth/github' element={renderRoute(<OAuth2Callback type='github'></OAuth2Callback>)} />
        <Route path='/oauth/discord' element={renderRoute(<OAuth2Callback type='discord'></OAuth2Callback>)} />
        <Route path='/oauth/oidc' element={renderRoute(<OAuth2Callback type='oidc'></OAuth2Callback>)} />
        <Route path='/oauth/linuxdo' element={renderRoute(<OAuth2Callback type='linuxdo'></OAuth2Callback>)} />
        <Route path='/oauth/:provider' element={renderRoute(<DynamicOAuth2Callback />)} />
        <Route
          path='/console/setting'
          element={
            <AdminRoute>
              {renderRoute(<Setting />)}
            </AdminRoute>
          }
        />
        <Route
          path='/console/personal'
          element={
            <PrivateRoute>
              {renderRoute(<PersonalSetting />)}
            </PrivateRoute>
          }
        />
        <Route
          path='/console/topup'
          element={
            <PrivateRoute>
              {renderRoute(<TopUp />)}
            </PrivateRoute>
          }
        />
        <Route
          path='/console/log'
          element={
            <PrivateRoute>
              {renderRoute(<Log />)}
            </PrivateRoute>
          }
        />
        <Route
          path='/console'
          element={
            <PrivateRoute>
              {renderRoute(<Dashboard />)}
            </PrivateRoute>
          }
        />
        <Route
          path='/console/midjourney'
          element={
            <PrivateRoute>
              {renderRoute(<Midjourney />)}
            </PrivateRoute>
          }
        />
        <Route
          path='/console/task'
          element={
            <PrivateRoute>
              {renderRoute(<Task />)}
            </PrivateRoute>
          }
        />
        <Route
          path='/pricing'
          element={
            pricingRequireAuth ? (
              <PrivateRoute>
                {renderRoute(<Pricing />)}
              </PrivateRoute>
            ) : (
              renderRoute(<Pricing />)
            )
          }
        />
        <Route
          path='/market'
          element={
            <PrivateRoute>
              <MarketplaceLegacyRedirect />
            </PrivateRoute>
          }
        />
        <Route path='/about' element={renderRoute(<About />)} />
        <Route path='/user-agreement' element={renderRoute(<UserAgreement />)} />
        <Route path='/privacy-policy' element={renderRoute(<PrivacyPolicy />)} />
        <Route path='/console/chat/:id?' element={renderRoute(<Chat />)} />
        {/* 方便使用chat2link直接跳转聊天... */}
        <Route
          path='/chat2link'
          element={
            <PrivateRoute>
              {renderRoute(<Chat2Link />)}
            </PrivateRoute>
          }
        />
        <Route path='*' element={<NotFound />} />
      </Routes>
    </SetupCheck>
  );
}

export default App;
