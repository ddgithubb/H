package com.horizontalmvcp;

import android.view.View;
import android.util.Log;

import com.facebook.react.bridge.Promise;
import com.facebook.react.bridge.ReactApplicationContext;
import com.facebook.react.bridge.ReactContextBaseJavaModule;
import com.facebook.react.bridge.ReactMethod;
import com.facebook.react.uimanager.IllegalViewOperationException;
import com.facebook.react.uimanager.NativeViewHierarchyManager;
import com.facebook.react.uimanager.ReactShadowNode;
import com.facebook.react.uimanager.UIBlock;
import com.facebook.react.uimanager.UIImplementation;
import com.facebook.react.uimanager.UIManagerModule;
import com.facebook.react.uimanager.UIManagerModuleListener;
import com.facebook.react.views.scroll.ReactHorizontalScrollView;
import com.facebook.react.views.view.ReactViewGroup;

import java.util.HashMap;

/**
 * Holds the required values for layoutUpdateListener.
 */
class HorizontalScrollViewUIHolders {
  static int prevFirstVisibleTop = 0;
  static View firstVisibleView = null;
  static int currentScrollX = 0;
}

public class MvcpHorizontalScrollViewManagerModule extends ReactContextBaseJavaModule {
  private final ReactApplicationContext reactContext;
  private HashMap<Integer, UIManagerModuleListener> uiManagerModuleListeners;

  MvcpHorizontalScrollViewManagerModule(ReactApplicationContext reactContext) {
    super(reactContext);
    this.reactContext = reactContext;
  }

  @Override
  public String getName() {
    return "MvcpHorizontalScrollViewManager";
  }

  @Override
  public void initialize() {
    super.initialize();
    this.uiManagerModuleListeners = new HashMap<>();
  }

  @ReactMethod
  public void enableMaintainVisibleContentPosition(final int viewTag, final int autoscrollToTopThreshold, final int minIndexForVisible, final Promise promise) {
    final UIManagerModule uiManagerModule = this.reactContext.getNativeModule(UIManagerModule.class);
    this.reactContext.runOnUiQueueThread(new Runnable() {
      @Override
      public void run() {
        try {
          final ReactHorizontalScrollView horizontalScrollView = (ReactHorizontalScrollView)uiManagerModule.resolveView(viewTag);
          final UIManagerModuleListener uiManagerModuleListener = new UIManagerModuleListener() {
            @Override
            public void willDispatchViewUpdates(final UIManagerModule uiManagerModule) {
              uiManagerModule.prependUIBlock(new UIBlock() {
                @Override
                public void execute(NativeViewHierarchyManager nativeViewHierarchyManager) {
                  ReactViewGroup mContentView = (ReactViewGroup)horizontalScrollView.getChildAt(0);
                  if (mContentView == null) return;

                  HorizontalScrollViewUIHolders.currentScrollX = horizontalScrollView.getScrollX();

                  for (int ii = minIndexForVisible; ii < mContentView.getChildCount(); ++ii) {
                    View subview = mContentView.getChildAt(ii);
                    if (subview.getLeft() >= HorizontalScrollViewUIHolders.currentScrollX) {
                      HorizontalScrollViewUIHolders.prevFirstVisibleTop = subview.getLeft();
                      HorizontalScrollViewUIHolders.firstVisibleView = subview;
                      break;
                    }
                  }
                }
              });
            }
          };

          UIImplementation.LayoutUpdateListener layoutUpdateListener = new UIImplementation.LayoutUpdateListener() {
            @Override
            public void onLayoutUpdated(ReactShadowNode root) {
              if (HorizontalScrollViewUIHolders.firstVisibleView == null) return;

              int deltaX = HorizontalScrollViewUIHolders.firstVisibleView.getLeft() - HorizontalScrollViewUIHolders.prevFirstVisibleTop;
              Log.d("MVCP", "deltaX" + deltaX);

              if (Math.abs(deltaX) > 1) {
                boolean isWithinThreshold = HorizontalScrollViewUIHolders.currentScrollX <= autoscrollToTopThreshold;
                horizontalScrollView.setScrollX(HorizontalScrollViewUIHolders.currentScrollX + deltaX);
                horizontalScrollView.fling(0);

                // If the offset WAS within the threshold of the start, animate to the start.
                if (isWithinThreshold) {
                  horizontalScrollView.smoothScrollTo(0, horizontalScrollView.getScrollY());
                }
              }
            }
          };

          uiManagerModule.getUIImplementation().setLayoutUpdateListener(layoutUpdateListener);
          uiManagerModule.addUIManagerListener(uiManagerModuleListener);
          int key = uiManagerModuleListeners.size() + 1;
          uiManagerModuleListeners.put(key, uiManagerModuleListener);
          promise.resolve(key);
        } catch(IllegalViewOperationException e) {
          promise.reject(e);
        }
      }
    });
  }

  @ReactMethod
  public void disableMaintainVisibleContentPosition(int key, Promise promise) {
    try {
      if (key >= 0) {
        final UIManagerModule uiManagerModule = this.reactContext.getNativeModule(UIManagerModule.class);
        uiManagerModule.removeUIManagerListener(uiManagerModuleListeners.remove(key));
        uiManagerModule.getUIImplementation().removeLayoutUpdateListener();
      }
      promise.resolve(null);
    } catch (Exception e) {
      promise.resolve(-1);
    }
  }
}
